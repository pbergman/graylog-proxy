package net

import (
	"errors"

	"github.com/pbergman/logger"
)

type ConnPoolInterface interface {
	// Close will close all connection
	// and all the open channels
	Close()
	// Wait will block till all workers
	// are finished (most times this is
	// when queue is empty/finished)
	Wait()
	// Start will start the workers in
	// the background to catch new
	// package forwards
	Start(workers int)
	// Push will create a non-blocking
	// queue item that can be later check
	Push(data []byte, id []byte) *ConnQueueItem
	// will write the given bytes to
	// the remote and block until we
	// we get some feedback
	Write(data []byte) (int, error)
}

func NewConnPool(noClientAuth bool, tries int, address *GraylogHost, ca, crt, pem string, logger logger.LoggerInterface) (ConnPoolInterface, error) {
	switch network := address.GetNetwork(); network {
	case "tcp", "tcp4", "tcp6":
		if !noClientAuth && address.IsSecure() {
			return NewTcpTlsConnPool(tries, address, ca, crt, pem, logger)
		} else {
			return NewTcpConnPool(tries, address, logger)
		}
	case "http", "https":
		if !noClientAuth && address.IsSecure() {
			return NewHttpsConnPool(tries, address, ca, crt, pem, logger)
		} else {
			return NewHttpConnPool(tries, address, logger)
		}
	default:
		return nil, errors.New("unsupported network provided '" + network + "'")
	}
}
