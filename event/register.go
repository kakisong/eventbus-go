package event

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"reflect"
	"runtime"
	"sync"
)

var (
	registerMu sync.RWMutex
	listeners  = make(map[string]map[string]interface{})
)

// 监听注册
func Register(handlerFunc interface{}) error {
	registerMu.Lock()
	defer registerMu.Unlock()
	funcType := reflect.TypeOf(handlerFunc)
	// 参数数量校验
	if err := checkFunc(funcType); err != nil {
		return err
	}
	paramName := generateParameterName(funcType)
	funcName := getFunctionName(handlerFunc)
	funcMap, ok := listeners[paramName]
	if ok {
		_, ok := funcMap[funcName]
		// 重复注册校验
		if ok {
			logrus.Errorf("方法 %s 已经被注册", funcName)
			return errors.New("该方法已被注册")
		}
		funcMap[funcName] = handlerFunc
		listeners[paramName] = funcMap
		logrus.Infof("方法 %s 监听注册成功", funcName)
	} else {
		logrus.Infof("方法 %s 监听注册成功", funcName)
		listeners[paramName] = map[string]interface{}{funcName: handlerFunc}
	}
	return nil
}

// 检测参数个数是否合法
func checkFunc(t reflect.Type) error {
	if t.Kind() != reflect.Func {
		return errors.New("只可使用方法进行监听")
	}
	if t.NumIn() < 1 {
		return errors.New("非标准方法，参数不可为空")
	}
	if t.NumIn() > 1 {
		return errors.New("非标准方法，只可拥有一个参数")
	}
	return nil
}

// 获取参数类型名
func generateParameterName(t reflect.Type) string {
	in := t.In(0)
	return in.PkgPath() + "." + in.Name()
}

// 获取函数名
func getFunctionName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
