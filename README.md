## 引言

最近项目中需要用到观察者模式来实现一些逻辑，如某些操作的数据变更会影响到同项目中另一模块的数据。

使用Java时可以用 Google Guava 中的 EventBus 来轻松实现，但是在go中暂无相似类库（可能有我不知道）。

由于是较为轻量级的应用，不想引入MQ这些外部实现，于是想到封装一个简单的实现。

文中的代码都放在 [https://github.com/kakisong/eventbus-go](https://github.com/kakisong/eventbus-go)

## 设计

由于goroutine天生的优势，所以在消息的转发过程会非常方便。

所以这里简单分为两个步骤：

- 监听函数的注册
- 接收到消息回调各个对应的监听函数

## 实现

### 监听函数的注册

由于我们想实现的是根据接收到的消息类型来决定消息转发的对应函数，这样使用起来就很方便，只需要确定监听的消息类型即可注册使用。

如：

现在有两种消息类型

- `EventTypeA`
- `EventTypeB`
四个函数
- `func1(EventTypeA)`
- `func2(EventTypeB)`
- `func3(EventTypeB)`
- `func4(EventTypeB, AnotherParam)`

那么我们现在需要实现的几个条件：

1. 接收到EventTypeB的时候需要回调 `func2`与`func3`
2. 接收到EventTypeB的时候只分别回调一次 `func2`与`func3`
3. 接收到EventTypeB的时候不回调`func1`与`func4`

上代码

```
var (
    // 注册锁
	registerMu sync.RWMutex
    // 监听函数存放处
	listeners  = make(map[string]map[string]reflect.Value)
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
		funcMap[funcName] = reflect.ValueOf(handlerFunc)
		listeners[paramName] = funcMap
		logrus.Infof("方法 %s 监听注册成功", funcName)
	} else {
		logrus.Infof("方法 %s 监听注册成功", funcName)
		listeners[paramName] = map[string]reflect.Value{funcName: reflect.ValueOf(handlerFunc)}
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
```

简单说明

1. 要将不符合我们条件的监听函数过滤调，我们的函数有切只能有一个参数，返回值可以不做控制，需要的话，可以控制返回值，我觉得其实这里的监听函数不应该有返回值
2. 获取全路径名与函数唯一入参的全路径名
3. 将监听函数的反射对象放入`listeners`，其实go的反射效率不高，所以这里存放反射后的对象

### 监听并转发消息

完成了监听函数的注册之后，接下来就是对发送过来的消息进行处理

- 定义一个普通的channel，缓冲大小为100

```
var msgChan = make(chan interface{}, 100)
```

- 简单的消息发送方法

```
// 发送消息
func Send(msg interface{}) {
	msgChan <- msg
}
```

- 消息监听方法

```
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
```

- 启动监听

```
func init() {
	go listen(msgChan)
}
```

这里直接放在init()方法中，使用时只要import我们的包就可以启动监听

## 测试

放上测试代码

```
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
```

测试结果

```
=== RUN   TestSend
time="2020-06-12T22:30:58+08:00" level=info msg="方法 github.com/kakisong/eventbus-go.PrintFunc 监听注册成功"
PrintFuncRun: code=0, msg=测试消息
--- PASS: TestSend (1.00s)
PASS
```

## 总结

这里我们利用了简洁的goroutine来实现消息的监听与消费，对外暴露的只有两个方法，我们不用去定义topic，不用去处理通道，由消息类型来控制回调的监听函数，在项目中的轻量级使用应该是开箱即用的。

- 函数注册 `Register`
- 消息发送 `Send`
所以使用起来也是十分简单方便的
