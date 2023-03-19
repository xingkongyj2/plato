package prpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hardcore-os/plato/common/prpc/discov/plugin"

	"github.com/bytedance/gopkg/util/logger"

	"github.com/hardcore-os/plato/common/prpc/discov"
	serverinterceptor "github.com/hardcore-os/plato/common/prpc/interceptor/server"
	"google.golang.org/grpc"
)

type RegisterFn func(*grpc.Server)

// 保存了gateway的grpc启动信息和etcd的注册信息
type PServer struct {
	//注意：没有写变量名，就可以使用便捷访问变量：PServer.serviceName
	serverOptions
	//函数数组，默认只有一个注册函数
	//RegisterFn：gRPC服务端注册函数，用来启动grpc服务
	registers []RegisterFn
	//todo：grpc什么设置
	interceptors []grpc.UnaryServerInterceptor
}

/* gateway向etcd中注册服务时，需要的一些信息。 其中包括了grpc的启动端口
 */
type serverOptions struct {
	//向etcd注册时的服务名
	serviceName string
	//grpc的启动地址和端口
	ip   string
	port int
	//gateway的连接情况统计信息
	weight int
	health bool
	//注册服务实例
	d discov.Discovery
}

type ServerOption func(opts *serverOptions)

// WithServiceName set serviceName
func WithServiceName(serviceName string) ServerOption {
	return func(opts *serverOptions) {
		opts.serviceName = serviceName
	}
}

// WithIP set ip
func WithIP(ip string) ServerOption {
	return func(opts *serverOptions) {
		opts.ip = ip
	}
}

// WithPort set port
func WithPort(port int) ServerOption {
	return func(opts *serverOptions) {
		opts.port = port
	}
}

// WithWeight set weight
func WithWeight(weight int) ServerOption {
	return func(opts *serverOptions) {
		opts.weight = weight
	}
}

// WithHealth set health
func WithHealth(health bool) ServerOption {
	return func(opts *serverOptions) {
		opts.health = health
	}
}

// opts中装了好几个函数，每个函数用来对一个指标量赋值
// 将gateway grpc的：服务名、IP、端口传进来了
func NewPServer(opts ...ServerOption) *PServer {
	opt := serverOptions{}
	//遍历函数，对opt赋值
	for _, o := range opts {
		//函数调用
		o(&opt)
	}

	if opt.d == nil {
		//根据配置文件中的ip地址创建一个ectd服务
		dis, err := plugin.GetDiscovInstance()
		if err != nil {
			panic(err)
		}

		opt.d = dis
	}

	return &PServer{
		opt,
		make([]RegisterFn, 0),
		make([]grpc.UnaryServerInterceptor, 0),
	}
}

// RegisterService ...
// eg :
//
//	p.RegisterService(func(server *grpc.Server) {
//	    test.RegisterGreeterServer(server, &Server{})
//	})
//
// 入参是一个函数
func (p *PServer) RegisterService(register ...RegisterFn) {
	//p.registers是一个数组，放的是register这个函数
	p.registers = append(p.registers, register...)
}

// RegisterUnaryServerInterceptor 注册自定义拦截器，例如限流拦截器或者自己的一些业务自定义拦截器
func (p *PServer) RegisterUnaryServerInterceptor(i grpc.UnaryServerInterceptor) {
	p.interceptors = append(p.interceptors, i)
}

// 启动grpc和etcd
func (p *PServer) Start(ctx context.Context) {
	//设置gateway在etcd中的kv。
	//key：配置中的服务名
	//value：grpc启动信息还有这个gateway的状态信息
	//所以只需要知道gateway在etcd中写入的key，就能拿到gateway的grpc信息和状态信息
	service := discov.Service{
		//etcd服务名：service_name: "plato.access.gateway"
		Name: p.serviceName,
		Endpoints: []*discov.Endpoint{
			{
				//etcd服务名
				ServerName: p.serviceName,
				//grpc ip+port
				IP:   p.ip,
				Port: p.port,
				//gateway的连接情况统计信息
				Weight: p.weight,
				Enable: true,
			},
		},
	}
	// ==========grpc start==========
	// 加载中间件
	interceptors := []grpc.UnaryServerInterceptor{
		serverinterceptor.RecoveryUnaryServerInterceptor(),
		serverinterceptor.TraceUnaryServerInterceptor(),
		serverinterceptor.MetricUnaryServerInterceptor(p.serviceName),
	}
	interceptors = append(interceptors, p.interceptors...)
	//创建gRPC服务器
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	//使用的是一个回调函数。可以开启多个grpc处理函数，默认只加了一个
	for _, register := range p.registers {
		//在gRPC服务端注册服务
		register(s)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", p.ip, p.port))
	if err != nil {
		panic(err)
	}
	//启动grpc服务
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	// ==========grpc end==========

	// etcd服务注册：将数据写入到etcd中
	p.d.Register(ctx, &service)

	logger.Info("start PRCP success")

	c := make(chan os.Signal, 1)
	//监听如下几个信号
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		sig := <-c
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			s.Stop()
			p.d.UnRegister(ctx, &service)
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}

}
