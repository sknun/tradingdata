package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-querystring/query"
	"github.com/sknun/tradingdata/pkg/model"
)

type BaseKlineBatchResponse struct {
	// 返回值 200 成功
	Code int `json:"code"`
	// 返回数据
	Data *[]model.KlineBatchResponse `json:"data,omitempty"`
	// 返回消息
	Message string `json:"message,omitempty"`
	// 返回错误消息 如果为空参考返回消息
	Error string `json:"error,omitempty"`
}

type BaseKlineResponse struct {
	// 返回值 200 成功
	Code int `json:"code"`
	// 返回数据
	Data *model.KlineResponse `json:"data,omitempty"`
	// 返回消息
	Message string `json:"message,omitempty"`
	// 返回错误消息 如果为空参考返回消息
	Error string `json:"error,omitempty"`
}

/*
操作控盘 偏移量模式
*/
func ControlOffsetAdd(host string, data model.OffsetRequest) (*model.Response, error) {
	return doGetRequest(host, "/api/control/offset", data)
}

/*
操作控盘 固定值模式
*/
func ControlFixedAdd(host string, data model.FixedRequest) (*model.Response, error) {
	return doGetRequest(host, "/api/control/fixed", data)
}

/*
操作控盘 画线模式
*/
func ControlDrawCreate(host string, data model.CreateDrawRequest) (*model.Response, error) {
	return doGetRequest(host, "/api/control/create_draw", data)
}

/*
操作控盘 画线模式
*/
func ControlDrawAdd(host string, data model.DrawRequest) (*model.Response, error) {
	return doGetRequest(host, "/api/control/draw", data)
}

/*
结束控盘
symbol 产品(格式:平台-产品)
*/
func ControlOver(host string, data model.OverRequest) (*model.Response, error) {
	return doGetRequest(host, "/api/control/over", data)
}

/*
删除控盘
*/
func ControlDelete(host string, data model.DeleteRequest) (*model.Response, error) {
	return doGetRequest(host, "/api/control/delete", data)
}

/*
获取历史k线
symbol 产品(格式:平台-产品)
kline_timestamp_end 截止时间
count 数量
resolution 颗度
*/

func HistoryKline(host, userID, token, symbol, kline_timestamp_end, count, resolution any) (*model.KlineResponse, error) {
	var tmp BaseKlineResponse
	if err := checkSymbolRes(symbol, resolution); err != nil {
		return nil, err
	}
	uri := fmt.Sprintf(
		"%v/api/kline_history?user_id=%v&token=%v&symbol=%v&count=%v&resolution=%v&kline_timestamp_end=%v",
		host,
		userID,
		token,
		symbol,
		count,
		resolution,
		kline_timestamp_end,
	)
	b, err := httpClientGet(uri)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 200 {
		return nil, fmt.Errorf("message:%v,error:%v", tmp.Message, tmp.Error)
	}
	return tmp.Data, nil
}

/*
批量获取历史k线
symbols 产品(格式:平台-产品) 多个用","分开
kline_timestamp_end 截止时间
count 数量
resolution 颗度
*/

func HistoryKlineBatch(host, userID, token, symbols, kline_timestamp_end, count, resolution any) (*[]model.KlineBatchResponse, error) {
	var tmp BaseKlineBatchResponse
	if err := checkSymbolRes(symbols, resolution); err != nil {
		return nil, err
	}
	uri := fmt.Sprintf(
		"%v/api/kline_history_batch?user_id=%v&token=%v&symbols=%v&count=%v&resolution=%v&kline_timestamp_end=%v",
		host,
		userID,
		token,
		symbols,
		count,
		resolution,
		kline_timestamp_end,
	)

	b, err := httpClientGet(uri)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 200 {
		return nil, fmt.Errorf("message:%v,error:%v", tmp.Message, tmp.Error)
	}
	return tmp.Data, err
}

// 请求数据源链接
func ConnectGrpc(host, userID, token, grpcAddr, grpcPort, grpcToken, symbols any) error {
	strSymbols, ok1 := symbols.(string)
	if !ok1 || strings.TrimSpace(strSymbols) == "" {
		return errors.New("缺少参数信息或参数类型错误")
	}
	uri := fmt.Sprintf(
		"%v/api/connection_delivery?user_id=%v&token=%v&your_sever=%v:%v&your_token=%v&symbols=%v",
		host,
		userID,
		token,
		grpcAddr,
		grpcPort,
		grpcToken,
		symbols,
	)
	b, err := httpClientGet(uri)
	if err != nil {
		return err
	}
	var res model.Response
	err = json.Unmarshal(b, &res)
	if err != nil {
		return err
	}
	if res.Code != 200 {
		return fmt.Errorf("服务器返回失败结果: %s,%s", res.Message, res.Error)
	}
	return nil
}

func checkSymbolRes(symbol, resolution any) error {
	strSymbol, ok1 := symbol.(string)
	strRes, ok2 := resolution.(string)
	if !ok1 || !ok2 || strings.TrimSpace(strSymbol) == "" || strings.TrimSpace(strRes) == "" {
		return errors.New("缺少参数信息或参数类型错误")
	}
	return nil
}

func httpClientGet(uri string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return []byte(""), err
	}
	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}
	return body, nil
}

func doGetRequest(host, path string, data any) (*model.Response, error) {
	var res model.Response

	values, err := query.Values(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %v", err)
	}

	uri := fmt.Sprintf("%v%v?%v", host, path, values.Encode())
	fmt.Println(uri)
	b, err := httpClientGet(uri)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func StringToUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}
