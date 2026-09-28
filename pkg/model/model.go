package model

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
	Platurl string `json:"platurl"`
	// 产品名
	Symbol string `json:"symbol"`
	// K线列表
	List []KlineData `json:"list"`
}

type KlineResponse struct {
	// 下次取历史k线的时间戳
	NextTimestamp int64 `json:"next_timestamp"`
	// K线列表
	List []KlineData `json:"list"`
}

type Response struct {
	// 返回值 200 成功
	Code int `json:"code"`
	// 返回数据
	Data any `json:"data"`
	// 返回消息
	Message string `json:"message"`
	// 返回错误消息 如果为空参考返回消息
	Error string `json:"error"`
}

type ConfigRequest struct {
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
}

type ListRequest struct {
	// 页码 可空 默认为1
	Page int `url:"page"`
	// 每页显示数量 可空 默认值由项目启动参数决定
	PageSize int `url:"page_size"`
	// 控盘ID
	ID uint `form:"id"`
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
	// 产品(格式:平台-产品)
	Symbol string `url:"symbol"`
}

type OffsetRequest struct {
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
	// 产品(格式:平台-产品)
	Symbol string `url:"symbol"`
	// 操作类型 immediate 瞬变模式 linear 渐变模式
	ModeType string `url:"mode_type"`
	// 偏移量
	Offset string `url:"offset"`
	// 偏移方向 1: up(涨) 2: down(跌)
	OffsetDirection string `url:"offset_direction"`
	// 开始时间戳(10位) 预设开始时间必须要大于当前时间 如果不设置不要提交此参数
	StartTime *string `url:"start_time,omitempty"`
	// 结束时间戳(10位) 不能大于用户最大允许时间 如果不设置不要提交此参数
	EndTime *string `url:"end_time,omitempty"`
	// 爬坡期时长(秒)
	RiseTime int64 `url:"rise_time"`
	// 回落期时长(秒)
	FallTime int64 `url:"fall_time"`
}

type FixedRequest struct {
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
	// 产品(格式:平台-产品)
	Symbol string `url:"symbol"`
	// 操作类型 immediate 瞬变模式 linear 渐变模式
	ModeType string `url:"mode_type"`
	// 大盘值 必须大于0
	FixedPrice string `url:"fixed_price"`
	// 开始时间戳(10位) 预设开始时间必须要大于当前时间 如果不设置不要提交此参数
	StartTime *string `url:"start_time,omitempty"`
	// 结束时间戳(10位) 不能大于用户最大允许时间 如果不设置不要提交此参数
	EndTime *string `url:"end_time,omitempty"`
	// 爬坡期时长(秒)
	RiseTime int64 `url:"rise_time"`
	// 回落期时长(秒)
	FallTime int64 `url:"fall_time"`
	// 上影线最小浮动值 如果不设置不要提交此参数
	UpperShadowMin *string `url:"upper_shadow_min,omitempty"`
	// 上影线最大浮动值 如果不设置不要提交此参数
	UpperShadowMax *string `url:"upper_shadow_max,omitempty"`
	// 下影线最小浮动值 如果不设置不要提交此参数
	LowerShadowMin *string `url:"lower_shadow_min,omitempty"`
	// 下影线最大浮动值 如果不设置不要提交此参数
	LowerShadowMax *string `url:"lower_shadow_max,omitempty"`
}

type DrawRequest struct {
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
	// 生成的画线ID
	ID string `url:"id"`
	// 上影线最小浮动值 如果不设置不要提交此参数
	UpperShadowMin *string `url:"upper_shadow_min,omitempty"`
	// 上影线最大浮动值 如果不设置不要提交此参数
	UpperShadowMax *string `url:"upper_shadow_max,omitempty"`
	// 下影线最小浮动值 如果不设置不要提交此参数
	LowerShadowMin *string `url:"lower_shadow_min,omitempty"`
	// 下影线最大浮动值 如果不设置不要提交此参数
	LowerShadowMax *string `url:"lower_shadow_max,omitempty"`
}

type CreateDrawRequest struct {
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
	// 产品(格式:平台-产品)
	Symbol string `url:"symbol"`
	// 爬坡期时长(秒) 必须大于59秒
	RiseTime int64 `url:"rise_time"`
	// 峰值期时长(秒) 必须大于59秒
	PeakTime int64 `url:"peak_time"`
	// 回落期时长(秒) 必须大于59秒
	FallTime int64 `url:"fall_time"`
	// 大盘值 FixedPrice和SpikeFactor必须传一项
	FixedPrice *string `url:"fixed_price,omitempty"`
	// 插针幅度 FixedPrice和SpikeFactor必须传一项
	SpikeFactor *string `url:"spike_factor,omitempty"`
}

type OverRequest struct {
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
	// 产品(格式:平台-产品)
	Symbol string `url:"symbol"`
}

type DeleteRequest struct {
	// 用户ID
	UserID uint `url:"user_id"`
	// token
	Token string `url:"token"`
	// 控盘ID 单条删除时只使用ID
	ID *uint `url:"id,omitempty"`
	// 产品(格式:平台-产品)
	Symbol *string `url:"symbol,omitempty"`
	// 删除范围: current-当前产品; all-所有产品
	Scope *string `url:"scope,omitempty"`
}

// 生成画线控盘的返回数据结构
type CreateDrawResponse struct {
	// 生成画线ID
	ID string `json:"id"`
	// K线列表
	List []KlineData `json:"list"`
}

func String(v string) *string { return &v }
