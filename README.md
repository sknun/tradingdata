# tradingdata Go SDK

用于查询历史K线、接收实时K线与盘口，以及调用控盘接口。需要 Go 1.25.1 或更高版本。

## 安装

```bash
go get github.com/sknun/tradingdata
```

## 接收实时行情

准备服务地址、用户ID、用户token、订阅产品，以及用于接收行情的公网IP或域名、端口和接收token。接收端口需要允许数据服务访问；使用端口映射时，外部端口应与SDK监听端口一致。

```go
package main

import (
    "log"

    "github.com/sknun/tradingdata"
)

func main() {
    client, err := tradingdata.NewClient(
        "https://api.example.com", // 服务地址
        "1",                       // 用户ID
        "YOUR_USER_TOKEN",         // 用户token
        "203.0.113.10",            // 客户程序可被访问的IP或域名，不带协议和端口
        "51015",                   // 接收端口
        "YOUR_RECEIVER_TOKEN",     // 自定义接收token
        "al-BTCUSDT,al-ETHUSDT",    // 平台-产品标识，多个用英文逗号分隔
        60,                        // 心跳超时，秒
        30,                        // 检查间隔，秒
        1000,                      // K线接收缓冲数量
        1000,                      // 盘口接收缓冲数量
    )
    if err != nil {
        log.Fatal(err)
    }
    client.SetLogger(func(message string) { log.Print(message) })

    go func() {
        for kline := range client.GetKlineChan() {
            // 在这里处理实时K线。
            _ = kline
        }
    }()
    go func() {
        for orderBook := range client.GetObChan() {
            // 在这里处理实时盘口。
            _ = orderBook
        }
    }()

    // Run启动接收服务并持续阻塞，应在独立进程或协程中运行。
    if err := client.Run(); err != nil {
        log.Fatal(err)
    }
}
```

每个进程只初始化一个客户端，`Run`只启动一次。心跳超时为60–1800秒，检查间隔为30–900秒，超时时间至少是检查间隔的两倍。

K线和盘口通道需要持续消费；缓冲满时会丢弃新行情并记录日志，不补发。`SetLogger`接入客户自己的日志系统，未设置时不输出SDK日志。

## 历史K线

以下示例使用已初始化的client：

```go
res, err := client.HistoryKline("al-BTCUSDT", 0, 100, "1m")
if err != nil {
    // 请求或解析失败。
    log.Print(err)
    return
}
if res.Code != 200 {
    // 服务返回业务失败，查看Message和Error。
    log.Print(res.Message, ": ", res.Error)
    return
}
if res.Data != nil {
    for _, item := range res.Data.List {
        // item包含时间、开高低收、成交量和成交额。
        _ = item
    }
    // 下一页查询使用res.Data.NextTimestamp作为截止时间。
}
```

批量查询：

```go
res, err := client.HistoryKlineBatch("al-BTCUSDT,al-ETHUSDT", 0, 100, "1m")
```

截止时间使用秒级时间戳，0表示最新数据；数量为1–200。支持周期：`1m、5m、15m、30m、1h、4h、D、W、M`。批量结果的`res.Data`为切片指针，每项包含平台、产品和K线列表。

## 控盘接口

请求和返回类型均位于 `github.com/sknun/tradingdata/pkg/model`。请求中的用户ID和token由客户端自动填写。

| 方法 | 请求类型 | 返回类型 |
| --- | --- | --- |
| ControlConfig | ConfigRequest | BaseControlConfigResponse |
| ControlList | ListRequest | BaseControlListResponse |
| ControlOffsetAdd | OffsetRequest | Response |
| ControlFixedAdd | FixedRequest | Response |
| ControlDrawCreate | CreateDrawRequest | BaseCreateDrawResponse |
| ControlDrawAdd | DrawRequest | Response |
| ControlOver | OverRequest | Response |
| ControlDelete | DeleteRequest | Response |

以上方法均返回`(*返回类型, error)`。例如：

```go
res, err := client.ControlList(model.ListRequest{
    Symbol:   "al-BTCUSDT",
    Page:     1,
    PageSize: 20,
})
```

请求字段的含义和可选条件见model中的注释。控盘请求的时间及阶段时长使用秒；控盘列表返回的时间及阶段时长使用毫秒，金额使用字符串保留精度。结束、删除接口返回成功表示操作已受理，后续处理可能尚未完成。

## 返回值与Swagger

所有查询和控盘方法保留服务返回的`Code、Data、Message、Error`。Go `error`表示请求或解析失败；业务是否成功需判断`res.Code == 200`。有固定格式的数据可通过`res.Data`直接访问。

客户HTTP接口可以直接返回res，Swagger引用SDK公开类型即可。例如历史K线接口：

```go
import sdkmodel "github.com/sknun/tradingdata/pkg/model"

// @Success 200 {object} sdkmodel.BaseKlineResponse
```

生成文档时启用依赖解析：

```bash
swag init --parseDependency
```
