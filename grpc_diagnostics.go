package tradingdata

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

var grpcInstanceID = fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())

// 日志仍交给SetLogger指定的实例；只记状态变化，不逐条打印行情或正常心跳。
func grpcLog(level, message string, fields map[string]any) {
	// 调用方可能保留心跳上下文，日志不得修改它。
	copyFields := make(map[string]any, len(fields)+3)
	for k, v := range fields {
		copyFields[k] = v
	}
	fields = copyFields
	fields["sdk_instance"] = grpcInstanceID
	fields["level"] = level
	fields["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	fields["msg"] = message
	b, _ := json.Marshal(fields)
	Logger(string(b))
}

type transportKey struct{}
type transportInfo struct{ ID, Remote, Local string }
type grpcDiagnostics struct {
	next atomic.Uint64
	boot string
}

func newGRPCDiagnostics() *grpcDiagnostics {
	return &grpcDiagnostics{boot: grpcInstanceID}
}
func (d *grpcDiagnostics) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, transportKey{}, transportInfo{
		ID: fmt.Sprintf("%s/%d", d.boot, d.next.Add(1)), Remote: info.RemoteAddr.String(), Local: info.LocalAddr.String()})
}
func (*grpcDiagnostics) HandleConn(ctx context.Context, event stats.ConnStats) {
	switch event.(type) {
	case *stats.ConnBegin:
		grpcLog("info", "SDK传输连接建立", rpcLogFields(ctx))
	case *stats.ConnEnd:
		grpcLog("info", "SDK传输连接结束", rpcLogFields(ctx))
	}
}
func (*grpcDiagnostics) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context { return ctx }
func (*grpcDiagnostics) HandleRPC(context.Context, stats.RPCStats)                       {}

func rpcLogFields(ctx context.Context) map[string]any {
	fields := make(map[string]any)
	md, _ := metadata.FromIncomingContext(ctx)
	for key, field := range map[string]string{
		"x-trading-connection-id": "connection_id", "x-trading-stream-id": "stream_id", "x-trading-user-id": "userID",
	} {
		if v := md.Get(key); len(v) > 0 {
			fields[field] = v[0]
		}
	}
	if v, ok := ctx.Value(transportKey{}).(transportInfo); ok {
		fields["transport_id"] = v.ID
		fields["peer"] = v.Remote
		fields["local_addr"] = v.Local
	} else if p, ok := peer.FromContext(ctx); ok {
		fields["peer"] = p.Addr.String()
	}
	return fields
}

// streamReceipt只有接收该流的goroutine写，记录最后已接收帧，不等同于客户业务已经消费。
type streamReceipt struct {
	started                      time.Time
	received, delivered, dropped uint64
	kind, platform, symbol       string
	seq                          uint64
	lastReceived                 time.Time
	lastDropLog                  time.Time
}

func (r *streamReceipt) fields(ctx context.Context) map[string]any {
	f := rpcLogFields(ctx)
	if last := lastHeartbeatReceipt.Load(); last != nil {
		f["last_heartbeat_at"] = last.at.UTC().Format(time.RFC3339Nano)
		f["last_heartbeat_connection_id"] = last.fields["connection_id"]
	}
	f["duration_ms"] = time.Since(r.started).Milliseconds()
	f["received"] = r.received
	f["enqueued"] = r.delivered
	f["dropped"] = r.dropped
	f["last_message_type"] = r.kind
	f["last_platform"] = r.platform
	f["last_symbol"] = r.symbol
	f["last_seq"] = r.seq
	if !r.lastReceived.IsZero() {
		f["last_received_at"] = r.lastReceived.UTC().Format(time.RFC3339Nano)
	}
	f["kline_queue_len"] = len(klineChan)
	f["kline_queue_cap"] = cap(klineChan)
	f["orderbook_queue_len"] = len(obChan)
	f["orderbook_queue_cap"] = cap(obChan)
	return f
}
func (r *streamReceipt) drop(ctx context.Context) {
	r.dropped++
	// 持续满队列最多每10秒一条，结束时再记录准确累计数，避免逐帧刷屏。
	if r.lastDropLog.IsZero() || time.Since(r.lastDropLog) >= 10*time.Second {
		grpcLog("error", "SDK接收通道已满，丢弃行情", r.fields(ctx))
		r.lastDropLog = time.Now()
	}
}
func logRPCError(ctx context.Context, message string, err error) {
	f := rpcLogFields(ctx)
	f["error"] = err.Error()
	f["grpc_code"] = status.Code(err).String()
	grpcLog("error", message, f)
}

// 保留最后一次成功心跳的连接标识，超时和重新登记时可追到同一连接。
type heartbeatReceipt struct {
	fields map[string]any
	at     time.Time
}

var lastHeartbeatReceipt atomic.Pointer[heartbeatReceipt]

func heartbeatLogFields() map[string]any {
	f := make(map[string]any)
	if last := lastHeartbeatReceipt.Load(); last != nil {
		maps.Copy(f, last.fields)
	}
	return f
}
