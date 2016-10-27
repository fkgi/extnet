package extnet

import (
	"log"
	"os"
	"testing"
)

const (
	localAddr  = "127.0.0.1:10000"
	remoteAddr = "127.0.0.1:10010"
	testString = "this is test"
)

func TestMain(m *testing.M) {
	Notificator = func(e error) { log.Println(e) }
	// ここにテストの初期化処理
	code := m.Run()
	// ここでテストのお片づけ
	os.Exit(code)
}
