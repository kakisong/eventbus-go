package eventbus_go

import (
	"fmt"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	Register(PrintFunc)
	Send(TestEvent{
		Code: 0,
		Msg:  "测试消息",
	})
	time.Sleep(1 * time.Second)
}

func PrintFunc(event TestEvent) {
	fmt.Printf("PrintFuncRun: code=%v, msg=%s\n", event.Code, event.Msg)
}
