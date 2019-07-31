package net

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/pbergman/logger"
)

type TcpTlsConnPool struct {
	address *GraylogHost
	config  *tls.Config
	connPool
}

func (c *TcpTlsConnPool) bind(conn *net.Conn) (err error) {
	dialer := new(net.Dialer)
	dialer.KeepAlive = c.KeepAlive
	dialer.Timeout = c.Timeout
	*conn, err = tls.DialWithDialer(dialer, c.address.GetNetwork(), c.address.GetHost(), c.config)
	return
}

func (c *TcpTlsConnPool) Start(workers int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.start(workers, c.bind)
}

func NewTcpTlsConnPool(tries int, address *GraylogHost, ca, crt, pem string, logger *logger.Logger) (ConnPoolInterface, error) {
	buf, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("loaded ca root certificate: '%s'", ca))
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(buf)
	pair, err := tls.LoadX509KeyPair(crt, pem)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("loaded certificate: '%s'", crt))
	logger.Debug(fmt.Sprintf("loaded private key: '%s'", pem))
	return &TcpTlsConnPool{
		address: address,
		config: &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{pair},
		},
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
