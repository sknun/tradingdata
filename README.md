# 数据中心客户端程序

### git@cyr.slmail.me

## 使用示例

```go
// 初始化客户端
client, err := NewClient(host, userID, token, grpcAddr, grpcPort, grpcToken, symbols, heartbeatTimeout, checkInterval, klineChanSize, obChanSize)
if err != nil {
    return fmt.Errorf("Client 初始化失败: %w", err)
}

// 注入日志函数 func printLog(l string) { log.Println(l) }
client.SetLogger(printLog)

// 获取k线
client.HistoryKline()

// 批量获取k线
client.HistoryKlineBatch()

// 获取并循环读取实时数据通道
for v := range client.GetKlineChan() {}
for v := range client.GetObChan() {}

// 启动实时数据获取
err = client.Run()
if err != nil {
    return fmt.Errorf("启动grpc失败: %w", err)
}
```

## gRPC断连诊断

客户程序必须通过 `client.SetLogger(...)` 将SDK日志保存到自己的gRPC接收日志实例；默认回调不输出，不能仅更新SDK却不接入日志。新增诊断日志为JSON，`level`只使用info/error，仍兼容现有 `func(string)` 回调，不引入新日志依赖。

- 新版dist通过metadata传入 `connection_id`、`stream_id`、`userID`，SDK日志原样记录；`sdk_instance`区分SDK进程，`transport_id`关联同一物理连接上的RPC。metadata仅用于诊断，不改变鉴权或行情格式。
- 记录服务启动、传输连接建立/结束、鉴权拒绝、每条流开始/首帧/结束、心跳超时与向core重新登记的结果。正常心跳和行情不逐条打印。
- 流结束包含最后接收时间、消息类型、产品、序号、接收数、成功入SDK通道数及丢弃数；入通道不等于客户业务已消费。满通道的重复错误最多每流每10秒一次，结束时给出准确累计数。
- `peer_half_closed`表示接收到了对端的流结束；`recv_failed`记录原始gRPC状态及context错误；发生接收处理panic时先记录再重新panic，保持原行为。传输连接结束本身不说明是谁主动关闭，需结合流记录和dist关闭原因。
- 心跳超时日志保留最后一次成功心跳的连接ID；登记请求返回成功仅表示core接口应答，不代表dist已连接。SDK服务重启看新的 `sdk_instance`；SIGKILL、断电等无机会写日志的退出需要宿主机或容器记录，不能当作已证明的网络故障。

跨端排查先匹配ID，再看本端关闭原因、SDK接收错误和最后消息；不要只凭两台机器日志的先后时间判断因果。两端时钟应同步。

本地离线验证：`go test -race ./... -run '^TestDiagnostic'`。现有 `TestClient_*` 是需要外部配置的调用示例，本命令不执行这些示例。

## HTTP接口返回类型

供调用方使用的HTTP方法保留core的 `code、data、message、error`。网络、超时、参数编码或JSON解析失败时返回Go `error`；core的业务失败码不转换为Go错误，由调用方判断。

| 方法 | 返回值 |
| --- | --- |
| HistoryKline | `(*model.BaseKlineResponse, error)` |
| HistoryKlineBatch | `(*model.BaseKlineBatchResponse, error)` |
| ControlConfig | `(*model.BaseControlConfigResponse, error)` |
| ControlList | `(*model.BaseControlListResponse, error)` |
| ControlDrawCreate | `(*model.BaseCreateDrawResponse, error)` |
| ControlOffsetAdd、ControlFixedAdd、ControlDrawAdd、ControlOver、ControlDelete | `(*model.Response, error)` |

`ConnectGrpc`仅用于SDK内部登记，保留原来的error返回和业务码判断，不向调用方提供响应。

返回类型均在 `pkg/model/model.go`。有固定数据格式的接口通过 `res.Data` 访问明确类型；无固定数据格式的操作以 `json.RawMessage` 保留core的原始data。

```go
res, err := client.ControlDrawCreate(request)
if err != nil {
    // 请求或解析失败。
    return err
}
// res.Code、res.Message、res.Error均保留core返回值。
// 调用方自行判断res.Code；成功且res.Data非nil时可访问res.Data.ID、res.Data.List。
```

调用方可以直接将res作为HTTP响应返回。Swagger按完整响应类型声明：

```go
import sdkmodel "github.com/sknun/tradingdata/pkg/model"

// 控盘列表接口：
// @Success 200 {object} sdkmodel.BaseControlListResponse

// 生成画线接口：
// @Success 200 {object} sdkmodel.BaseCreateDrawResponse

// 历史K线接口：
// @Success 200 {object} sdkmodel.BaseKlineResponse
```

生成文档时使用 `swag init --parseDependency`。若调用方只转发res.Data并使用自己的响应外壳，再写 `resp.Response{data=sdkmodel.KlineResponse}`。

控盘金额字段使用字符串保留精度。控盘列表的开始／结束时间和阶段时长为毫秒；提交参数仍按请求字段注释使用秒。core结束／删除应答成功不代表dist已完成后续清理。
