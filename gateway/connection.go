package gateway

import (
	"net"
	"sync/atomic"
)

var nextConnID uint64 // 全局的分配变量值

//每个连接分配一个连接对象
type connection struct {
	//每个连接的id
	id uint64 // 进程级别的生命周期
	//客户端的文件描述符
	fd   int
	e    *epoller
	conn *net.TCPConn
}

func NewConnection(conn *net.TCPConn) *connection {
	connID := atomic.AddUint64(&nextConnID, 1)
	return &connection{
		id:   connID,
		fd:   socketFD(conn),
		conn: conn,
	}
}
func (c *connection) Close() {
	ep.tables.Delete(c.id)
	if c.e != nil {
		c.e.fdToConnTable.Delete(c.fd)
	}
	err := c.conn.Close()
	panic(err)
}

func (c *connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *connection) BindEpoller(e *epoller) {
	c.e = e
}
