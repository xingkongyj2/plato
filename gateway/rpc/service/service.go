package service

import (
	context "context"
)

const (
	DelConnCmd = 1 // DelConn
	PushCmd    = 2 // push
)

type CmdContext struct {
	Ctx *context.Context
	//gateway连接发生错误发送CancelConnCmd，连接没问题就发送SendMsgCmd
	Cmd    int32
	ConnID uint64
	//客户端发来的整个协议数据
	Payload []byte
}

// 启动gprc之前这个类会与grpc注册绑定
// 我用这个类去实现gateway.proto文件中定义的服务
type Service struct {
	//gateway启动grpc服务之后，state会调用。传递一些数据保存到CmdChannel，然后在cmdHandler函数中去消费
	CmdChannel chan *CmdContext
}

// 服务端主动断开连接
func (s *Service) DelConn(ctx context.Context, gr *GatewayRequest) (*GatewayResponse, error) {
	//为什么要新建一个context？
	//因为入参中grpc传进来的context当这个函数结束后就会被释放了，如果还有其他协程用这个context，就到导致这些协程被关闭
	//TODO返回的也是一个空的context
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx:    &c,
		Cmd:    DelConnCmd,
		ConnID: gr.ConnID,
	}
	return &GatewayResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}

// 下行消息
func (s *Service) Push(ctx context.Context, gr *GatewayRequest) (*GatewayResponse, error) {
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx:     &c,
		Cmd:     PushCmd,
		ConnID:  gr.ConnID,
		Payload: gr.GetData(),
	}
	return &GatewayResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}
