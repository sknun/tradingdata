package tradingdata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/sknun/tradingdata/internal/api"
	"github.com/sknun/tradingdata/pkg/grpcproto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var (
	heartbeatTimeout time.Duration
	checkInterval    time.Duration
	lastSeen         atomic.Value
	klineChan        chan *grpcproto.Kline
	obChan           chan *grpcproto.OrderBook
	token            string
	isStarted        int32 // 0 表示未启动，1 表示已成功启动
)

func (c *Client) GetKlineChan() <-chan *grpcproto.Kline {
	return klineChan
}
func (c *Client) GetObChan() <-chan *grpcproto.OrderBook {
	return obChan
}

type server struct {
	grpcproto.UnimplementedServiceServer
}

func (s *server) PushData(stream grpc.ClientStreamingServer[grpcproto.Request, emptypb.Empty]) (result error) {
	ctx := stream.Context()
	receipt := streamReceipt{started: time.Now()}
	grpcLog("info", "SDK推送流开始接收", rpcLogFields(ctx))
	reason := "handler_return"
	defer func() {
		f := receipt.fields(ctx)
		if recovered := recover(); recovered != nil {
			f["panic"] = fmt.Sprint(recovered)
			grpcLog("error", "SDK接收处理panic", f)
			panic(recovered) // 记录后保持原来的panic行为，不静默吞掉程序错误。
		}
		f["reason"] = reason
		f["grpc_code"] = status.Code(result).String()
		level := "info"
		if result != nil {
			level = "error"
			f["error"] = result.Error()
		}
		if err := ctx.Err(); err != nil {
			f["context_error"] = err.Error()
		}
		grpcLog(level, "SDK推送流结束", f)
	}()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			reason = "peer_half_closed"
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			reason = "recv_failed"
			return err
		}
		receipt.received++
		receipt.lastReceived = time.Now()
		switch x := req.GetPayload().(type) {
		case *grpcproto.Request_Kline:
			receipt.kind = "kline"
			receipt.platform = x.Kline.Plat
			receipt.symbol = x.Kline.Symbol
			receipt.seq = x.Kline.Seq
			select {
			case klineChan <- x.Kline:
				receipt.delivered++
			default:
				receipt.drop(ctx)
			}
		case *grpcproto.Request_Ob:
			receipt.kind = "orderbook"
			receipt.platform = x.Ob.Plat
			receipt.symbol = x.Ob.Symbol
			receipt.seq = x.Ob.Seq
			select {
			case obChan <- x.Ob:
				receipt.delivered++
			default:
				receipt.drop(ctx)
			}
		default:
			f := rpcLogFields(ctx)
			f["payload_type"] = fmt.Sprintf("%T", req.GetPayload())
			grpcLog("error", "SDK收到空或未知payload，忽略该帧", f)
			continue
		}
		if receipt.received == 1 {
			grpcLog("info", "SDK收到推送流首帧", receipt.fields(ctx))
		}
	}
}

// 20s推送一次心跳
func (s *server) Heartbeat(ctx context.Context, req *grpcproto.HeartbeatRequest) (*grpcproto.HeartbeatResponse, error) {
	_ = req
	now := time.Now()
	lastSeen.Store(now)
	fields := rpcLogFields(ctx)
	previous := lastHeartbeatReceipt.Swap(&heartbeatReceipt{fields: fields, at: now})
	if id, ok := fields["connection_id"]; ok && (previous == nil || previous.fields["connection_id"] != id) {
		grpcLog("info", "SDK收到新连接心跳", fields)
	}
	return &grpcproto.HeartbeatResponse{ServerTime: time.Now().Unix()}, nil
}

// 拦截器：校验唯一客户端的 token
func checkAuth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid connection")
	}
	t := md.Get("authorization")
	if len(t) == 0 || t[0] != token {
		return status.Error(codes.Unauthenticated, "invalid connection.")
	}
	return nil
}

func unaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	_ = info
	if err := checkAuth(ctx); err != nil {
		logRPCError(ctx, "SDK RPC鉴权失败", err)
		return nil, err
	}

	return handler(ctx, req)
}

func streamInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	_ = info
	if err := checkAuth(ss.Context()); err != nil {
		logRPCError(ss.Context(), "SDK推送流鉴权失败", err)
		return err
	}
	return handler(srv, ss)
}

// 心跳监测看守进程
func startHeartbeatMonitor(ontimeoutFunc func()) {
	// 初始化全局时间
	lastSeen.Store(time.Now())

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		// 取出最后活跃时间
		last, ok := lastSeen.Load().(time.Time)
		if !ok {
			continue
		}

		if time.Since(last) > heartbeatTimeout {
			fields := heartbeatLogFields()
			fields["last_heartbeat_at"] = last.UTC().Format(time.RFC3339Nano)
			fields["elapsed_ms"] = time.Since(last).Milliseconds()
			fields["timeout_ms"] = heartbeatTimeout.Milliseconds()
			grpcLog("error", "SDK心跳超时，触发重新登记", fields)
			// 先行重置时间，防止在 api.ConnectGrpc() 耗时太长时，定时器触发下一次循环造成连续触发
			lastSeen.Store(time.Now())
			// 触发重连回调
			ontimeoutFunc()
		}
	}
}

// Run 启动服务
func (c *Client) Run() error {
	if !atomic.CompareAndSwapInt32(&isStarted, 0, 1) {
		return errors.New("包级全局服务正在运行中，禁止重复调用 Run()")
	}
	token = c.GrpcToken
	heartbeatTimeout = time.Duration(c.HeartbeatTimeout) * time.Second
	checkInterval = time.Duration(c.CheckInterval) * time.Second

	lis, err := net.Listen("tcp", ":"+c.GrpcPort)
	if err != nil {
		atomic.StoreInt32(&isStarted, 0)
		return fmt.Errorf("unary RPC Failed to listen, error: %w", err)
	}

	go c.onTimeout()

	// 启动常驻的心跳监控协程
	go startHeartbeatMonitor(c.onTimeout)

	opts := []grpc.ServerOption{
		grpc.StatsHandler(newGRPCDiagnostics()),
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	}

	grpcServer := grpc.NewServer(opts...)
	grpcproto.RegisterServiceServer(grpcServer, &server{})

	grpcLog("info", "SDK gRPC服务启动", map[string]any{"userID": c.UserID, "listen_port": c.GrpcPort, "heartbeat_timeout_seconds": c.HeartbeatTimeout})
	if err := grpcServer.Serve(lis); err != nil {
		grpcLog("error", "SDK gRPC服务异常退出", map[string]any{"error": err.Error()})
		atomic.StoreInt32(&isStarted, 0)
		return fmt.Errorf("unary RPC Failed to serve, error: %w", err)
	}
	return nil
}

// 超时触发数据源重接grpc
func (c *Client) onTimeout() {
	started := time.Now()
	fields := heartbeatLogFields()
	fields["userID"] = c.UserID
	fields["receiver_addr"] = c.GrpcAddr + ":" + c.GrpcPort
	grpcLog("info", "SDK请求core登记接收地址", fields)
	if err := api.ConnectGrpc(c.Host, c.UserID, c.Token, c.GrpcAddr, c.GrpcPort, c.GrpcToken, c.Symbols); err != nil {
		grpcLog("error", "SDK登记接收地址失败", map[string]any{"userID": c.UserID, "error": err.Error(), "elapsed_ms": time.Since(started).Milliseconds()})
		return
	}
	grpcLog("info", "SDK登记请求返回成功，等待dist实际连接", map[string]any{"userID": c.UserID, "elapsed_ms": time.Since(started).Milliseconds()})
}
