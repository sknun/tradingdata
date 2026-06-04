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
	"github.com/sknun/tradingdata/internal/grpcproto"

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

func (s *server) PushData(stream grpc.ClientStreamingServer[grpcproto.Request, emptypb.Empty]) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return err
		}

		payload := req.GetPayload()
		if payload == nil {
			Logger("【警告】收到空payload(oneof未设置),已忽略该帧,请通知数据中心排查")
			continue
		}

		switch x := payload.(type) {
		case *grpcproto.Request_Kline:
			select {
			case klineChan <- x.Kline:
			default:
				Logger(fmt.Sprintf("【警告】Kline 通道已满,丢弃当前帧: %s,请合理优化读取消息函数", x.Kline.Code))
			}
		case *grpcproto.Request_Ob:
			select {
			case obChan <- x.Ob:
			default:
				Logger(fmt.Sprintf("【警告】Ob 通道已满,丢弃当前帧: %s,请合理优化读取消息函数", x.Ob.Code))
			}
		default:
			Logger(fmt.Sprintf("【警告】收到未知payload类型: %T,请通知数据中心排查", x))
		}
	}
}

// 20s推送一次心跳
func (s *server) Heartbeat(ctx context.Context, req *grpcproto.HeartbeatRequest) (*grpcproto.HeartbeatResponse, error) {
	_ = req
	lastSeen.Store(time.Now())
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
			Logger("【系统提示】唯一的客户端断开或失去响应，正在重置计时并尝试激活连接...")
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

	Logger("正在首次通知数据源连接 gRPC...")
	go c.onTimeout()

	// 启动常驻的心跳监控协程
	go startHeartbeatMonitor(c.onTimeout)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	}

	grpcServer := grpc.NewServer(opts...)
	grpcproto.RegisterServiceServer(grpcServer, &server{})

	Logger(fmt.Sprintf("gRPC 独占数据接收服务端已启动，正在监听端口: %s ...", c.GrpcPort))
	if err := grpcServer.Serve(lis); err != nil {
		atomic.StoreInt32(&isStarted, 0)
		return fmt.Errorf("unary RPC Failed to serve, error: %w", err)
	}
	return nil
}

// 超时触发数据源重接grpc
func (c *Client) onTimeout() {
	if err := api.ConnectGrpc(c.Host, c.UserID, c.Token, c.GrpcAddr, c.GrpcPort, c.GrpcToken, c.Codes); err != nil {
		Logger(fmt.Sprintf("通知数据源连接 gRPC 失败: %v", err))
	}
}
