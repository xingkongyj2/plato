package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/hardcore-os/plato/common/config"
	"github.com/hardcore-os/plato/common/prpc"
	"github.com/hardcore-os/plato/common/tcp"
	"github.com/hardcore-os/plato/gateway/rpc/client"
	"github.com/hardcore-os/plato/gateway/rpc/service"
	"google.golang.org/grpc"
)

var cmdChannel chan *service.CmdContext

// RunMain 启动网关服务
func RunMain(path string) {
	config.Init(path)
	//设置监听端口：8900
	//注意：只是设置监听端口，并不是开始监听。accept才是开启监听端口
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: config.GetGatewayTCPServerPort()})
	if err != nil {
		log.Fatalf("StartTCPEPollServer err:%s", err.Error())
		panic(err)
	}
	//创建协程池
	initWorkPoll()
	//初始化epoll
	initEpoll(ln, runProc)

	fmt.Println("-------------im gateway stated------------")
	/**
	grpc和etcd启动的整体逻辑：
	    1.初始化grpc和etcd启动信息
	    2.配置grpc的注册函数
	    3.启动grpc和etcd
	*/
	//设置gateway的grpc服务信息
	//创建一个空cmdChannel
	cmdChannel = make(chan *service.CmdContext, config.GetGatewayCmdChannelNum())
	//1.初始化grpc和etcd启动信息
	s := prpc.NewPServer(
		prpc.WithServiceName(config.GetGatewayServiceName()),
		prpc.WithIP(config.GetGatewayServiceAddr()),
		prpc.WithPort(config.GetGatewayRPCServerPort()), prpc.WithWeight(config.GetGatewayRPCWeight()))
	fmt.Println(config.GetGatewayServiceName(), config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort(), config.GetGatewayRPCWeight())

	//2.配置grpc的注册函数
	//func(server *grpc.Server)这个入参函数等于2：
	//    1.s := grpc.NewServer()                  // 创建gRPC服务器
	//    2.pb.RegisterGreeterServer(s, &server{}) // 在gRPC服务端注册服务
	//    3.err = s.Serve(lis) //启动服务
	s.RegisterService(func(server *grpc.Server) {
		//创建的cmdChannel初始化gateway grpc的Service类
		//gateway 在gRPC服务端注册服务。注册grpc服务的时候，第二个参数是一个自定义类，这个类实现了grpc中的服务
		//当启动之后，有一个rpc请求过来的时候，就会创建一个service.Service类，然后调用函数。
		service.RegisterGatewayServer(server, &service.Service{CmdChannel: cmdChannel})
	})

	// 启动state rpc 客户端
	client.Init()
	// 启动 命令处理写协程
	go cmdHandler()
	//3.启动grpc和etcd
	s.Start(context.TODO())
}

/*
从epoll监听到数据后，就使用这个函数处理：gateway对数据不做处理，通过grpc传給state模块。这里除了使用grpc通信，还能使用本地socket通信。
c *connectio：epoll监听到数据后返回的连接
ep *epoller：就是epoll
*/
func runProc(c *connection, ep *epoller) {
	ctx := context.Background() // 起始的contenxt
	// step1: 读取一个完整的消息包
	dataBuf, err := tcp.ReadData(c.conn)
	if err != nil {
		// 如果读取conn时发现连接关闭，则直接端口连接
		// 通知 state 清理掉意外退出的 conn的状态信息
		if errors.Is(err, io.EOF) {
			// 这步操作是异步的，不需要等到返回成功在进行，因为消息可靠性的保障是通过协议完成的而非某次cmd
			ep.remove(c)
			client.CancelConn(&ctx, getEndpoint(), c.id, nil)
		}
		return
	}
	//将函数給协程池
	err = wPool.Submit(func() {
		// step2:交给 state server rpc 处理
		//getEndpoint()就是IP地址和端口
		client.SendMsg(&ctx, getEndpoint(), c.id, dataBuf)
	})
	if err != nil {
		fmt.Errorf("runProc:err:%+v\n", err.Error())
	}
}

// 读取service.Service中cmdChannel的值
// state传递过来报文，在这里消费
func cmdHandler() {
	for cmd := range cmdChannel {
		// 异步提交到协池中完成发送任务
		switch cmd.Cmd {
		//服务器主动断开
		case service.DelConnCmd:
			wPool.Submit(func() { closeConn(cmd) })
		//下行消息
		case service.PushCmd:
			wPool.Submit(func() { sendMsgByCmd(cmd) })
		default:
			panic("command undefined")
		}
	}
}
func closeConn(cmd *service.CmdContext) {
	if connPtr, ok := ep.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		conn.Close()
	}
}

// 将state发来的下行消息发送给客户端
func sendMsgByCmd(cmd *service.CmdContext) {
	if connPtr, ok := ep.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		dp := tcp.DataPgk{
			Len:  uint32(len(cmd.Payload)),
			Data: cmd.Payload,
		}
		tcp.SendData(conn.conn, dp.Marshal())
	}
}

func getEndpoint() string {
	return fmt.Sprintf("%s:%d", config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort())
}
