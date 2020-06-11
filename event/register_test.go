package event

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type TestEvent struct {
	Code int
	Msg  string
}

func TestRegister(t *testing.T) {
	Convey("监听注册方法", t, func() {
		Convey("注册成功 --> 新的消息体监听方法注册", func() {
			err := Register(OneParamFunc)
			So(err, ShouldBeNil)
		})
		Convey("注册成功 --> 已有消息体监听方法注册 --> with receiver", func() {
			e := TestEvent{}
			err := Register(e.OneParamFunc)
			So(err, ShouldBeNil)
		})
		Convey("注册失败 --> 参数为空", func() {
			err := Register(ZeroParamFunc)
			So(err, ShouldNotBeNil)
		})
		Convey("注册失败 --> 参数超长", func() {
			err := Register(TwoParamsFunc)
			So(err, ShouldNotBeNil)
		})
		Convey("注册失败 --> 重复注册", func() {
			err := Register(OneParamFunc)
			err = Register(OneParamFunc)
			So(err, ShouldNotBeNil)
		})
	})
}

func ZeroParamFunc() {
}

func OneParamFunc(TestEvent) {
}

func (t *TestEvent) OneParamFunc(TestEvent) {
}

func TwoParamsFunc(TestEvent, TestEvent) {
}
