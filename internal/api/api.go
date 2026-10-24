package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/sknun/tradingdata/pkg/model"
)

/*
操作控盘 控盘配置
*/
func ControlConfig(host string, data model.ConfigRequest) (*model.BaseControlConfigResponse, error) {
	return doGetRequest[model.ControlConfigResponse](host, "/api/control/config", data)
}

/*
操作控盘 控盘列表
*/
func ControlList(host string, data model.ListRequest) (*model.BaseControlListResponse, error) {
	return doGetRequest[model.ControlListResponse](host, "/api/control/list", data)
}

/*
操作控盘 偏移量模式
*/
func ControlOffsetAdd(host string, data model.OffsetRequest) (*model.Response, error) {
	return doGetRequest[json.RawMessage](host, "/api/control/offset", data)
}

/*
操作控盘 固定值模式
*/
func ControlFixedAdd(host string, data model.FixedRequest) (*model.Response, error) {
	return doGetRequest[json.RawMessage](host, "/api/control/fixed", data)
}

/*
操作控盘 画线模式
*/
func ControlDrawCreate(host string, data model.CreateDrawRequest) (*model.BaseCreateDrawResponse, error) {
	return doGetRequest[model.CreateDrawResponse](host, "/api/control/create_draw", data)
}

/*
操作控盘 画线模式
*/
func ControlDrawAdd(host string, data model.DrawRequest) (*model.Response, error) {
	return doGetRequest[json.RawMessage](host, "/api/control/draw", data)
}

/*
结束控盘
symbol 产品(格式:平台-产品)
*/
func ControlOver(host string, data model.OverRequest) (*model.Response, error) {
	return doGetRequest[json.RawMessage](host, "/api/control/over", data)
}

/*
删除控盘
*/
func ControlDelete(host string, data model.DeleteRequest) (*model.Response, error) {
	return doGetRequest[json.RawMessage](host, "/api/control/delete", data)
}

/*
获取历史k线
symbol 产品(格式:平台-产品)
kline_timestamp_end 截止时间
count 数量
resolution 颗度
*/

func HistoryKline(host, userID, token, symbol, kline_timestamp_end, count, resolution any) (*model.BaseKlineResponse, error) {
	var tmp model.BaseKlineResponse
	if err := checkSymbolRes(symbol, resolution); err != nil {
		return nil, err
	}
	// 编码查询参数，避免token等字段中的特殊字符被当作URL分隔符。
	params := url.Values{}
	params.Set("user_id", fmt.Sprint(userID))
	params.Set("token", fmt.Sprint(token))
	params.Set("symbol", fmt.Sprint(symbol))
	params.Set("count", fmt.Sprint(count))
	params.Set("resolution", fmt.Sprint(resolution))
	params.Set("kline_timestamp_end", fmt.Sprint(kline_timestamp_end))
	uri := fmt.Sprintf("%v/api/kline_history?%s", host, params.Encode())
	b, err := httpClientGet(uri)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

/*
批量获取历史k线
symbols 产品(格式:平台-产品) 多个用","分开
kline_timestamp_end 截止时间
count 数量
resolution 颗度
*/

func HistoryKlineBatch(host, userID, token, symbols, kline_timestamp_end, count, resolution any) (*model.BaseKlineBatchResponse, error) {
	var tmp model.BaseKlineBatchResponse
	if err := checkSymbolRes(symbols, resolution); err != nil {
		return nil, err
	}
	// 编码查询参数，避免token等字段中的特殊字符被当作URL分隔符。
	params := url.Values{}
	params.Set("user_id", fmt.Sprint(userID))
	params.Set("token", fmt.Sprint(token))
	params.Set("symbols", fmt.Sprint(symbols))
	params.Set("count", fmt.Sprint(count))
	params.Set("resolution", fmt.Sprint(resolution))
	params.Set("kline_timestamp_end", fmt.Sprint(kline_timestamp_end))
	uri := fmt.Sprintf("%v/api/kline_history_batch?%s", host, params.Encode())
	b, err := httpClientGet(uri)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

// 请求数据源链接
func ConnectGrpc(host, userID, token, grpcAddr, grpcPort, grpcToken, symbols any) error {
	strSymbols, ok1 := symbols.(string)
	if !ok1 || strings.TrimSpace(strSymbols) == "" {
		return errors.New("缺少参数信息或参数类型错误")
	}
	// 接收token可能包含+、&等字符，编码后由core还原原值。
	params := url.Values{}
	params.Set("user_id", fmt.Sprint(userID))
	params.Set("token", fmt.Sprint(token))
	params.Set("your_sever", fmt.Sprintf("%v:%v", grpcAddr, grpcPort))
	params.Set("your_token", fmt.Sprint(grpcToken))
	params.Set("symbols", strSymbols)
	uri := fmt.Sprintf("%v/api/connection_delivery?%s", host, params.Encode())
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
	client := &http.Client{Timeout: 15 * time.Second}
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

// doGetRequest只处理请求和解析错误；core业务码及消息由调用方决定如何处理。
func doGetRequest[T any](host, path string, data any) (*model.BaseResponse[T], error) {
	var res model.BaseResponse[T]

	values, err := query.Values(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %v", err)
	}

	uri := fmt.Sprintf("%v%v?%v", host, path, values.Encode())
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
