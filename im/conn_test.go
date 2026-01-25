package im

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/xuning888/helloIMClient/internal/http"
	"github.com/xuning888/helloIMClient/pkg/logger"
	"github.com/xuning888/helloIMClient/protocol"
)

func dispatchMsg(frame *protocol.Frame) {

}

var seqs = atomic.Int32{}

func getSeq() int32 {
	add := seqs.Add(1)
	return add
}

func Test_Conn(t *testing.T) {
	logger.InitLogger()
	http.Init("http://127.0.0.1:8087", time.Second*3)
	ms := NewMsgSender(dispatchMsg)
	conn := NewHiConn(ms, getSeq)
	conn.Connect()
	for {
		select {}
	}
}
