package model

import (
	"encoding/json"
	"time"
)

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

// BaseResponse 保留core的完整业务响应；SDK不根据Code生成error。
type BaseResponse[T any] struct {
	// 返回码，200表示成功。
	Code int `json:"code"`
	// 具体业务数据，无数据的操作可能不返回此字段。
	Data *T `json:"data,omitempty"`
	// 返回消息。
	Message string `json:"message,omitempty"`
	// 具体失败原因。
	Error string `json:"error,omitempty"`
}

// BaseKlineResponse 是单个产品历史K线的完整HTTP响应。
type BaseKlineResponse = BaseResponse[KlineResponse]

// BaseKlineBatchResponse 是批量历史K线的完整HTTP响应。
type BaseKlineBatchResponse = BaseResponse[[]KlineBatchResponse]

// Response 用于没有固定Data结构的操作，保留core返回的原始Data。
type Response = BaseResponse[json.RawMessage]

// BaseControlConfigResponse 是控盘配置的完整响应。
type BaseControlConfigResponse = BaseResponse[ControlConfigResponse]

// BaseControlListResponse 是控盘列表的完整响应。
type BaseControlListResponse = BaseResponse[ControlListResponse]

// BaseCreateDrawResponse 是生成画线的完整响应。
type BaseCreateDrawResponse = BaseResponse[CreateDrawResponse]

// ControlConfigResponse 是ControlConfig返回的数据。
type ControlConfigResponse struct {
	// 支持的K线周期。
	SupportedResolutions []string `json:"supported_resolutions"`
	// 是否支持组请求。
	SupportsGroupRequest bool `json:"supports_group_request"`
	// 是否支持标记。
	SupportsMarks bool `json:"supports_marks"`
	// 是否支持时间轴标记。
	SupportsTimescaleMarks bool `json:"supports_timescale_marks"`
	// 是否支持服务器时间。
	SupportsTime bool `json:"supports_time"`
}

// ControlListResponse 是ControlList返回的分页数据。
type ControlListResponse struct {
	// 当前页码。
	Page int `json:"page"`
	// 每页数量。
	PageSize int `json:"page_size"`
	// 控盘列表。
	List []ControlInfo `json:"list"`
	// 总记录数。
	Total int64 `json:"total"`
}

// ControlInfo 对应core控盘列表的JSON字段。金额使用字符串保留原始精度。
type ControlInfo struct {
	// 控盘ID。
	ID uint `json:"id"`
	// 创建时间，RFC3339格式；未设置时为null。
	CreatedAt *time.Time `json:"created_at"`
	// 更新时间，RFC3339格式；未设置时为null。
	UpdatedAt *time.Time `json:"updated_at"`
	// [common] 平台ID
	PlatformID uint `json:"platform_id"`
	// [common] 产品ID
	SymbolID uint `json:"symbol_id"`
	// [common] 用户ID
	UserID uint `json:"user_id"`
	// [common] 控盘类型 offset 偏移量控盘 fixed 大盘值控盘 draw 画线控盘
	ControlType string `json:"control_type"`
	// [common] 控盘状态 0 未开始 1 进行中 2 结束
	ControlStatus uint8 `json:"control_status"`
	// [common] 控盘结束类型 1 自然结束 2 手动结束 3 下发失败 4 异常结束N
	ControlEndType uint8 `json:"control_end_type"`
	// [common] 备注
	Remark string `json:"remark"`
	// [common] 爬坡期时长(毫秒)
	RiseTime int64 `json:"rise_time"`
	// [common] 峰值期时长(毫秒)
	PeakTime int64 `json:"peak_time"`
	// [common] 回落期时长(毫秒)
	FallTime int64 `json:"fall_time"`
	// [common] 上影线最小浮动值
	UpperShadowMin string `json:"upper_shadow_min"`
	// [common] 上影线最大浮动值
	UpperShadowMax string `json:"upper_shadow_max"`
	// [common] 下影线最小浮动值
	LowerShadowMin string `json:"lower_shadow_min"`
	// [common] 下影线最大浮动值
	LowerShadowMax string `json:"lower_shadow_max"`
	// 峰值浮动最小值
	PeakMin string `json:"peak_min"`
	// 峰值浮动最大值
	PeakMax string `json:"peak_max"`
	// [common] 卖盘浮动最小值
	AsksPriceMin string `json:"asks_price_min"`
	// [common] 卖盘浮动最大值
	AsksPriceMax string `json:"asks_price_max"`
	// [common] 买盘浮动最小值
	BidsPriceMin string `json:"bids_price_min"`
	// [common] 买盘浮动最大值
	BidsPriceMax string `json:"bids_price_max"`
	// [common] 买卖盘阶梯价差最小值
	ObDiffMin string `json:"ob_diff_min"`
	// [common] 买卖盘阶梯价差最大值
	ObDiffMax string `json:"ob_diff_max"`

	// [draw] 爬坡期的命中概率最小值 取值范围为 [0, 1]
	RiseHitMin float64 `json:"rise_hit_min"`
	// [draw] 爬坡期的命中概率最大值 取值范围为 [0, 1]
	RiseHitMax float64 `json:"rise_hit_max"`
	// [draw] 回落期的命中概率最小值 取值范围为 [0, 1]
	FallHitMin float64 `json:"fall_hit_min"`
	// [draw] 回落期的命中概率最大值 取值范围为 [0, 1]
	FallHitMax float64 `json:"fall_hit_max"`
	// [draw] 爬坡期的命中后取值范围最小值
	RiseValueMin string `json:"rise_value_min"`
	// [draw] 爬坡期的命中后取值范围最大值
	RiseValueMax string `json:"rise_value_max"`
	// [draw] 回落期的命中后取值范围最小值
	FallValueMin string `json:"fall_value_min"`
	// [draw] 回落期的命中后取值范围最大值
	FallValueMax string `json:"fall_value_max"`
	// [draw] 峰值期的浮动值
	PeakFloat string `json:"peak_float"`

	// @Mode: offset fixed
	// 控盘模式 immediate 瞬变模式 linear 渐变模式
	ModeType string `json:"mode_type"`
	// @Mode: offset
	// 偏移量
	Offset string `json:"offset"`
	// @Mode: offset
	// 偏移方向 1: up(涨) 2: down(跌)
	OffsetDirection string `json:"offset_direction"`
	// @Mode: fixed
	// 大盘值
	FixedPrice string `json:"fixed_price"`
	// @Mode: offset fixed draw
	// 开始时间戳(13位) 预设开始时间必须要大于当前时间 留空为立即执行
	StartTime int64 `json:"start_time"`
	// @Mode: offset fixed draw
	// 结束时间戳(13位) 不能大于用户最大允许时间 留空为用户最大允许时间
	EndTime int64 `json:"end_time"`
	// 是否删除
	IsDelete bool `json:"is_delete"`
	// 配置来源 false为默认 true 用户
	ConfSource bool `json:"conf_source"`
	// @Mode: draw
	// 画线模式数据
	CandlesInfo json.RawMessage `json:"candles_info"`
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
	ID uint `url:"id"`
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
