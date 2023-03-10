package source

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/hardcore-os/plato/common/config"
	"github.com/hardcore-os/plato/common/discovery"
)

func Init() {
	//eventChan是一个全局变量，存放统计的数据
	eventChan = make(chan *Event)
	//开始监听
	ctx := context.Background()
	go DataHandler(&ctx)
	if config.IsDebug() {
		ctx := context.Background()
		testServiceRegister(&ctx, "7896", "node1")
		testServiceRegister(&ctx, "7897", "node2")
		testServiceRegister(&ctx, "7898", "node3")
	}
}

// 调用commom里面的discovery
// etcd服务发现
func DataHandler(ctx *context.Context) {
	//ip config是etcd中的服务发现，拿到gateway统计的数据，进行负载均衡
	//设置服务发现
	dis := discovery.NewServiceDiscovery(ctx)
	//协程结束的时候关闭dis
	defer dis.Close()
	//监听etcd，删除和增加会使用下面的两个函数来处理
	//闭包函数，添加节点
	setFunc := func(key, value string) { //key和value是序列化过的
		//反序列化value
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			//拿到统计数据event
			if event := NewEvent(ed); ed != nil {
				event.Type = AddNodeEvent
				//生产端：将监听到的数据放入通道
				eventChan <- event
			}
		} else {
			logger.CtxErrorf(*ctx, "DataHandler.setFunc.err :%s", err.Error())
		}
	}
	//删除节点
	delFunc := func(key, value string) {
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			if event := NewEvent(ed); ed != nil {
				event.Type = DelNodeEvent
				eventChan <- event
			}
		} else {
			logger.CtxErrorf(*ctx, "DataHandler.delFunc.err :%s", err.Error())
		}
	}
	//监听前缀（key）。前缀从配置中读取
	err := dis.WatchService(config.GetServicePathForIPConf(), setFunc, delFunc)
	if err != nil {
		panic(err)
	}
}
