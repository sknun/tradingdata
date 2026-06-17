package tradingdata

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/sknun/tradingdata/pkg/grpcproto"
)

type Client struct {
	Host             string // 数据中心地址
	UserID           string // 用户ID
	Token            string // 用户token
	GrpcAddr         string // 接收实时数据的grpc地址
	GrpcPort         string // 接收实时数据的grpc端口
	GrpcToken        string // 接收实时数据时验证的token
	Codes            string // 请求实时数据的产品列表,使用半角逗号分割(平台-产品标识)
	HeartbeatTimeout int    // 心跳超时时间(秒) 必须是检查的两倍时间以上
	CheckInterval    int    // 检查的频率(秒)
	KlineChanSize    int    // k线通道缓冲区大小 不建议小于100
	OBChanSize       int    // 盘口通道缓冲区大小 不建议小于100
}

var isClientCreated int32

func NewClient(
	host,
	userID,
	token,
	grpcAddr,
	grpcPort,
	grpcToken string,
	codes string,
	heartbeatTimeout,
	checkInterval,
	klineChanSize,
	obChanSize int,
) (*Client, error) {
	if !atomic.CompareAndSwapInt32(&isClientCreated, 0, 1) {
		return nil, errors.New("tradingdata 客户端已经初始化过，禁止重复调用 NewClient()")
	}
	if strings.TrimSpace(token) == "" {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, errors.New("token 不能为空")
	}
	if strings.TrimSpace(grpcToken) == "" {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, errors.New("grpcToken 不能为空")
	}
	u, err := url.Parse(host)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, fmt.Errorf("host 必须是一个合法的 https 地址，当前输入: %s", host)
	}
	uID, err := strconv.Atoi(userID)
	if err != nil || uID <= 0 {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, fmt.Errorf("userID 必须是大于 0 的数字，当前输入: %s", userID)
	}
	gPort, err := strconv.Atoi(grpcPort)
	if err != nil || gPort <= 0 || gPort > 65535 {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, fmt.Errorf("grpcPort 必须是合法的端口数字(1-65535)，当前输入: %s", grpcPort)
	}
	trimmedAddr := strings.TrimSpace(grpcAddr)
	if trimmedAddr == "" {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, errors.New("grpcAddr 不能为空")
	}
	if net.ParseIP(trimmedAddr) == nil && !strings.Contains(trimmedAddr, ".") && trimmedAddr != "localhost" {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, fmt.Errorf("grpcAddr 必须是合法的 IPv4/IPv6 地址或域名，当前输入: %s", grpcAddr)
	}
	if len(codes) < 1 {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, errors.New("产品列表 不能为空")
	}
	if err := checkCodes(codes); err != nil {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, err
	}
	if heartbeatTimeout < 60 || heartbeatTimeout > 1800 {
		heartbeatTimeout = 60
	}
	if checkInterval < 30 || checkInterval > 900 {
		checkInterval = 30
	}
	if heartbeatTimeout < checkInterval*2 {
		atomic.StoreInt32(&isClientCreated, 0)
		return nil, fmt.Errorf("heartbeatTimeout(%d) 必须严格大于等于 checkInterval(%d) 的 2 倍", heartbeatTimeout, checkInterval)
	}
	if klineChanSize < 100 {
		klineChanSize = 1000
	}
	if obChanSize < 100 {
		obChanSize = 1000
	}
	// 先初始化通道
	klineChan = make(chan *grpcproto.Kline, klineChanSize)
	obChan = make(chan *grpcproto.OrderBook, obChanSize)
	return &Client{
		Host:             host,
		UserID:           userID,
		Token:            token,
		GrpcAddr:         grpcAddr,
		GrpcPort:         grpcPort,
		GrpcToken:        grpcToken,
		Codes:            codes,
		HeartbeatTimeout: heartbeatTimeout,
		CheckInterval:    checkInterval,
		KlineChanSize:    klineChanSize,
		OBChanSize:       obChanSize,
	}, nil
}

func checkCodes(codes string) error {
	slice := strings.Split(codes, ",")
	seen := make(map[string]struct{})
	if len(slice) < 1 {
		return errors.New("产品列表不能为空")
	}

	for _, code := range slice {
		if strings.TrimSpace(code) == "" {
			return errors.New("产品列表中包含合法的空参数")
		}

		if _, exists := seen[code]; exists {
			return errors.New("产品列表中包含重复的数据: " + code)
		}

		seen[code] = struct{}{}
	}

	return nil
}
