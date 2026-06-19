package tradingdata

import (
	"github.com/sknun/tradingdata/internal/api"
	"github.com/sknun/tradingdata/pkg/klinestruct"
)

/*
获取历史k线
code 产品(平台-产品标识)
kline_timestamp_end 截止时间 0为取最新数据
count 数量 1-200之间
resolution 颗度 1m 5m 15m 30m 1h 4h D W M
*/
func (c *Client) HistoryKline(code, kline_timestamp_end, count, resolution any) (res *klinestruct.KlineResponse, err error) {
	return api.HistoryKline(c.Host, c.UserID, c.Token, code, kline_timestamp_end, count, resolution)
}

/*
批量获取历史k线
codes 产品(平台-产品标识) 多个用","分开
kline_timestamp_end 截止时间 0为取最新数据
count 数量 1-200之间
resolution 颗度 1m 5m 15m 30m 1h 4h D W M
*/
func (c *Client) HistoryKlineBatch(codes, kline_timestamp_end, count, resolution any) (res *[]klinestruct.KlineBatchResponse, err error) {
	return api.HistoryKlineBatch(c.Host, c.UserID, c.Token, codes, kline_timestamp_end, count, resolution)
}
