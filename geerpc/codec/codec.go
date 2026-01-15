package codec

import "io"

// package codec 存放和消息编解码相关的代码

type Header struct {
	// 服务名和方法名 通常与Go语言中的结构体和方法相映射
	ServiceMethod string // format "Service.Method"
	// 请求的序号 某个请求的id 用来区分不同的请求
	Seq           uint64 // sequence number chosen by client
	// 错误信息 客户端置为空 服务器如果发生错误 将错误信息置于Error中
	Error         string
}

// 抽象出对消息体进行编解码的接口Codec
// 抽象出接口是为了实现不同的Codec实例
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

// 抽象出Codec的构造函数
// 类似工厂模式
type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init()  {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}