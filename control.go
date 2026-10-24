package tradingdata

import (
	"github.com/sknun/tradingdata/internal/api"
	"github.com/sknun/tradingdata/pkg/model"
)

/*
操作控盘 控盘配置
*/
func (c *Client) ControlConfig(data model.ConfigRequest) (*model.BaseControlConfigResponse, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlConfig(c.Host, data)
}

/*
操作控盘 控盘列表
*/
func (c *Client) ControlList(data model.ListRequest) (*model.BaseControlListResponse, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlList(c.Host, data)
}

/*
操作控盘 偏移量模式
*/
func (c *Client) ControlOffsetAdd(data model.OffsetRequest) (*model.Response, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlOffsetAdd(c.Host, data)
}

/*
操作控盘 固定值模式
*/
func (c *Client) ControlFixedAdd(data model.FixedRequest) (*model.Response, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlFixedAdd(c.Host, data)
}

/*
操作控盘 画线模式生成
*/
func (c *Client) ControlDrawCreate(data model.CreateDrawRequest) (*model.BaseCreateDrawResponse, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlDrawCreate(c.Host, data)
}

/*
操作控盘 画线模式
*/
func (c *Client) ControlDrawAdd(data model.DrawRequest) (*model.Response, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlDrawAdd(c.Host, data)
}

/*
结束控盘
*/
func (c *Client) ControlOver(data model.OverRequest) (*model.Response, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlOver(c.Host, data)
}

/*
删除控盘
*/
func (c *Client) ControlDelete(data model.DeleteRequest) (*model.Response, error) {
	data.UserID = api.StringToUint(c.UserID)
	data.Token = c.Token
	return api.ControlDelete(c.Host, data)
}
