package net

import (
	"net"
	"time"

	"github.com/pbergman/logger"
)

type TcpConnPool struct {
	address *net.TCPAddr
	connPool
}

func (c *TcpConnPool) bind(conn *net.Conn) (err error) {
	*conn, err = net.DialTCP("tcp", nil, c.address)
	return
}

func (c *TcpConnPool) Start(workers int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.start(workers, c.bind)
}

func NewTcpConnPool(tries int, host *GraylogHost, logger *logger.Logger) (ConnPoolInterface, error) {
	if address, err := net.ResolveTCPAddr(host.GetNetwork(), host.GetHost()); err != nil {
		return nil, err
	} else {
		return &TcpConnPool{
			address: address,
			connPool: connPool{
				KeepAlive: 3 * time.Minute,
				Timeout:   1 * time.Minute,
				logger:    logger,
				connQueue: connQueue{
					tries: tries,
					queue: make(chan *ConnQueueItem, 10),
				},
			},
		}, nil
	}
}
