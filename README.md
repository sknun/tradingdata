# 数据中心客户端程序

### git@cyr.slmail.me

## 使用示例

```go
// 初始化客户端
client, err := NewClient(host, userID, token, grpcAddr, grpcPort, grpcToken, codes, heartbeatTimeout, checkInterval, klineChanSize, obChanSize)
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