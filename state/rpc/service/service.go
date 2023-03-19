package service

import (
	context "context"
)

const (
	CancelConnCmd = 1
	SendMsgCmd    = 2
)

type CmdContext struct {
	Ctx *context.Context
	//gateway连接发生错误发送CancelConnCmd，连接没问题就发送SendMsgCmd
	Cmd int32
	//发消息过来的gateway的grpc服务地址
	Endpoint string
	ConnID   uint64
	//客户端发来的整个协议数据
	Payload []byte
}

type Service struct {
	CmdChannel chan *CmdContext
}

func (s *Service) CancelConn(ctx context.Context, sr *StateRequest) (*StateResponse, error) {
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx:      &c,
		Cmd:      CancelConnCmd,
		ConnID:   sr.ConnID,
		Endpoint: sr.GetEndpoint(),
	}
	return &StateResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}

func (s *Service) SendMsg(ctx context.Context, sr *StateRequest) (*StateResponse, error) {
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx:      &c,
		Cmd:      SendMsgCmd,
		ConnID:   sr.ConnID,
		Endpoint: sr.GetEndpoint(),
		Payload:  sr.GetData(),
	}
	return &StateResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}
