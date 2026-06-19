package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sknun/tradingdata/pkg/klinestruct"
)

type BaseKlineBatchResponse struct {
	// 返回值 200 成功
	Code int `json:"code"`
	// 返回数据
	Data *[]klinestruct.KlineBatchResponse `json:"data,omitempty"`
	// 返回消息
	Message string `json:"message,omitempty"`
	// 返回错误消息 如果为空参考返回消息
	Error string `json:"error,omitempty"`
}

type BaseKlineResponse struct {
	// 返回值 200 成功
	Code int `json:"code"`
	// 返回数据
	Data *klinestruct.KlineResponse `json:"data,omitempty"`
	// 返回消息
	Message string `json:"message,omitempty"`
	// 返回错误消息 如果为空参考返回消息
	Error string `json:"error,omitempty"`
}

type ConnectResponse struct {
	// 返回值 200 成功
	Code int `json:"code"`
	// 返回数据
	Data any `json:"data,omitempty"`
	// 返回消息
	Message string `json:"message,omitempty"`
	// 返回错误消息 如果为空参考返回消息
	Error string `json:"error,omitempty"`
}

/*
获取历史k线
code 产品
kline_timestamp_end 截止时间
count 数量
resolution 颗度
*/

func HistoryKline(host, userID, token, code, kline_timestamp_end, count, resolution any) (*klinestruct.KlineResponse, error) {
	var tmp BaseKlineResponse
	if err := checkCodeRes(code, resolution); err != nil {
		return nil, err
	}
	uri := fmt.Sprintf(
		"%v/api/kline_history?user_id=%v&token=%v&code=%v&count=%v&resolution=%v&kline_timestamp_end=%v",
		host,
		userID,
		token,
		code,
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
	return tmp.Data, nil
}

/*
批量获取历史k线
codes 产品 多个用","分开
kline_timestamp_end 截止时间
count 数量
resolution 颗度
*/

func HistoryKlineBatch(host, userID, token, codes, kline_timestamp_end, count, resolution any) (*[]klinestruct.KlineBatchResponse, error) {
	var tmp BaseKlineBatchResponse
	if err := checkCodeRes(codes, resolution); err != nil {
		return nil, err
	}
	uri := fmt.Sprintf(
		"%v/api/kline_history_batch?user_id=%v&token=%v&codes=%v&count=%v&resolution=%v&kline_timestamp_end=%v",
		host,
		userID,
		token,
		codes,
		count,
		resolution,
		kline_timestamp_end,
	)

	b, err := httpClientGet(uri)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &tmp)
	return tmp.Data, err
}

// 请求数据源链接
func ConnectGrpc(host, userID, token, grpcAddr, grpcPort, grpcToken, codes any) error {
	strCode, ok1 := codes.(string)
	if !ok1 || strings.TrimSpace(strCode) == "" {
		return errors.New("缺少参数信息或参数类型错误")
	}
	uri := fmt.Sprintf(
		"%v/api/connection_delivery?user_id=%v&token=%v&your_sever=%v:%v&your_token=%v&codes=%v",
		host,
		userID,
		token,
		grpcAddr,
		grpcPort,
		grpcToken,
		codes,
	)
	b, err := httpClientGet(uri)
	if err != nil {
		return err
	}
	var res ConnectResponse
	err = json.Unmarshal(b, &res)
	if err != nil {
		return err
	}
	if res.Code != 200 {
		return fmt.Errorf("服务器返回失败结果: %s,%s", res.Message, res.Error)
	}
	return nil
}

func checkCodeRes(code, resolution any) error {
	strCode, ok1 := code.(string)
	strRes, ok2 := resolution.(string)
	if !ok1 || !ok2 || strings.TrimSpace(strCode) == "" || strings.TrimSpace(strRes) == "" {
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
