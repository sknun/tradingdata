package tradingdata

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"strings"
	"testing"
)

func TestClient_HistoryKline(t *testing.T) {
	tstr, _ := generateRandomSecret(80)
	grpcToken = gmd5(tstr)
	client, err := NewClient(host, userID, gentoken, grpcAddr, grpcPort, grpcToken, strings.Join(gencodes, ","), genheartbeatTimeout, gencheckInterval, genklineChanSize, genobChanSize)
	if err != nil {
		t.Fatalf("Client 初始化失败: %v", err)
	}
	result, err := client.HistoryKline("al-Silver", "0", "10", "1m")

	if err != nil {
		t.Fatalf("请求历史k线失败: %v", err)
	}

	t.Logf("直连测试成功！实际数据: %v", result)
}

func TestClient_HistoryKlineBatch(t *testing.T) {
	tstr, _ := generateRandomSecret(80)
	grpcToken = gmd5(tstr)
	client, err := NewClient(host, userID, gentoken, grpcAddr, grpcPort, grpcToken, strings.Join(gencodes, ","), genheartbeatTimeout, gencheckInterval, genklineChanSize, genobChanSize)
	if err != nil {
		t.Fatalf("Client 初始化失败: %v", err)
	}
	result, err := client.HistoryKlineBatch("al-Silver", "0", "10", "1m")

	if err != nil {
		t.Fatalf("请求批量历史k线失败: %v", err)
	}

	t.Logf("直连测试成功！实际数据: %v", result)
}

func TestClient_Grpc(t *testing.T) {
	tstr, _ := generateRandomSecret(80)
	grpcToken = gmd5(tstr)
	client, err := NewClient(host, userID, gentoken, grpcAddr, grpcPort, grpcToken, strings.Join(gencodes, ","), genheartbeatTimeout, gencheckInterval, genklineChanSize, genobChanSize)
	if err != nil {
		t.Fatalf("Client 初始化失败: %v", err)
	}
	client.SetLogger(printLog)
	go func() {
		for v := range client.GetKlineChan() {
			log.Println(v)
		}
	}()
	go func() {
		for v := range client.GetObChan() {
			log.Println(v)
		}
	}()
	err = client.Run()

	if err != nil {
		t.Fatalf("启动grpc失败: %v", err)
	}
}

func printLog(l string) { log.Println(l) }

func gmd5(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	md5Hash := hash.Sum(nil)
	return hex.EncodeToString(md5Hash)
}
