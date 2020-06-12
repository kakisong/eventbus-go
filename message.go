package eventbus_go

import "reflect"

var msgChan = make(chan interface{}, 100)

func init() {
	go listen(msgChan)
}

// 发送消息
func Send(msg interface{}) {
	msgChan <- msg
}

// 监听
func listen(msgChan <-chan interface{}) {
	for {
		select {
		case msg := <-msgChan:
			call(msg)
		default:
		}
	}
}

// 调用监听方法
func call(param interface{}) {
	registerMu.Lock()
	defer registerMu.Unlock()
	paramName := getParamName(param)
	funcMap, ok := listeners[paramName]
	if ok {
		for _, method := range funcMap {
			go method.Call([]reflect.Value{reflect.ValueOf(param)})
		}
	}
}

// 获取类型名
func getParamName(param interface{}) string {
	t := reflect.TypeOf(param)
	return t.PkgPath() + "." + t.Name()
}
