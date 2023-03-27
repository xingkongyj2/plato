package perf

import (
	"fmt"
	"net"
	"time"

	"github.com/hardcore-os/plato/common/sdk"
)

var (
	TcpConnNum int32
)

func RunMain() {
	for i := 0; i < int(TcpConnNum); i++ {
		fmt.Printf("连接服务数量：%d\n", i)
		time.Sleep(100)
		sdk.NewChat(net.ParseIP("127.0.0.1"), 8900, "logic", "1223", "123")
	}
}
