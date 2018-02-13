package net

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pbergman/logger"
)

type connPool struct {
	connQueue

	pool      []net.Conn
	lock      sync.Mutex
	logger    logger.LoggerInterface
	KeepAlive time.Duration
	Timeout   time.Duration
}

func (c *connPool) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, conn := range c.pool {
		if nil != conn {
			conn.Close()
		}
	}
	c.pool = nil
	close(c.queue)
}

func (c *connPool) start(workers int, bind func(*net.Conn) (err error)) {
	c.pool = make([]net.Conn, workers)
	for i := 0; i < workers; i++ {
		c.wg.Add(1)
		go c.process(c.pool[i], &c.wg, bind)
	}
}

func (c *connPool) process(conn net.Conn, wg *sync.WaitGroup, bind func(*net.Conn) (err error)) {
	defer wg.Done()
	for item := range c.queue {
		if nil == conn {
			if err := bind(&conn); err != nil {
				// on error just return/stop this and
				// close message because this should
				// be a connection error
				c.logger.Error(err)
				close(item.status)
				return
			}
		}
		n, err := conn.Write(item.data)
		c.logger.Info(fmt.Sprintf("[%X] written %d bytes to '%s'", item.id, n, conn.RemoteAddr().String()))
		if err != nil {
			c.logger.Error(fmt.Sprintf("[%X] %s", item.id, err.Error()))
			item.tries++
			item.error = append(item.error, err)
			if item.tries < 5 {
				c.queue <- item
			} else {
				c.logger.Alert(fmt.Sprintf("[%X] discarded message after %d retires", item.id, item.tries))
			}
			// on error just reset the connection this
			// could be a timeout or closed connection.
			conn.Close()
			conn = nil
		} else {
			close(item.status)
		}
	}
}
