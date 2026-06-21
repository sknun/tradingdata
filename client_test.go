package tradingdata

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"log"
	"math/big"
	"strings"
	"testing"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	host     = ""
	userID   = ""
	gentoken = ""
	grpcAddr = ""
	grpcPort = ""
)

var (
	grpcToken  string
	gensymbols = []string{
		"al-BTCUSDT",
		"al-ETHUSDT",
		"al-GOLD",
		"al-Silver",
		// "al-AAPL.US",
	}
	genheartbeatTimeout int = 60
	gencheckInterval    int = 30
	genklineChanSize    int = 1000
	genobChanSize       int = 1000
)

func TestClient_HistoryKline(t *testing.T) {
	tstr, _ := generateRandomSecret(80)
	grpcToken = gmd5(tstr)
	client, err := NewClient(host, userID, gentoken, grpcAddr, grpcPort, grpcToken, strings.Join(gensymbols, ","), genheartbeatTimeout, gencheckInterval, genklineChanSize, genobChanSize)
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
	client, err := NewClient(host, userID, gentoken, grpcAddr, grpcPort, grpcToken, strings.Join(gensymbols, ","), genheartbeatTimeout, gencheckInterval, genklineChanSize, genobChanSize)
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
	client, err := NewClient(host, userID, gentoken, grpcAddr, grpcPort, grpcToken, strings.Join(gensymbols, ","), genheartbeatTimeout, gencheckInterval, genklineChanSize, genobChanSize)
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

func generateRandomSecret(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[num.Int64()]
	}
	return string(result), nil
}
