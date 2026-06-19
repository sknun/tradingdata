package klinestruct

type KlineData struct {
	Time     int64   `json:"time"`     // K线时间
	Open     float64 `json:"open"`     // 开盘价
	Close    float64 `json:"close"`    // 收盘价
	High     float64 `json:"high"`     // 最高价
	Low      float64 `json:"low"`      // 最低价
	Volume   float64 `json:"volume"`   // 成交量
	Turnover float64 `json:"turnover"` // 成交额
}

type KlineBatchResponse struct {
	// 平台
	Platform string `json:"platform"`
	// 产品名
	Code string `json:"code"`
	// K线列表
	List []KlineData `json:"list"`
}

type KlineResponse struct {
	// 下次取历史k线的时间戳
	NextTimestamp int64 `json:"next_timestamp"`
	// K线列表
	List []KlineData `json:"list"`
}
