package net

import (
    "crypto/tls"
    "io/ioutil"
    "crypto/x509"
    "sync"
    "bytes"
    "net/http"

    "github.com/pbergman/logger"
)

type HttpsConnPool struct {
    config *tls.Config
    HttpConnPool
}

func (p HttpsConnPool) Start(workers int) {
    p.clients = make([]*http.Client, workers)
    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        p.clients[i] = &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: p.config,
            },
        }
        go p.process(p.clients[i], &p.wg)
    }
}

func NewHttpsConnPool(tries int, host *GraylogHost, ca, crt, pem string, logger logger.LoggerInterface) (ConnPoolInterface, error) {
    buf, err := ioutil.ReadFile(ca)
    if err != nil {
        return nil, err
    }
    certPool := x509.NewCertPool()
    certPool.AppendCertsFromPEM(buf)
    pair, err := tls.LoadX509KeyPair(crt, pem)
    if err != nil {
        return nil, err
    }
    return &HttpsConnPool{
        HttpConnPool: HttpConnPool{
            host: host,
            logger: logger,
            pool: &sync.Pool{
                New: func() interface{} {
                    return new(bytes.Buffer)
                },
            },
            connQueue: connQueue {
                tries: tries,
                queue: make(chan *ConnQueueItem, 10),
            },
        },
        config: &tls.Config{
            RootCAs:      certPool,
            Certificates: []tls.Certificate{pair},
        },
    }, nil
}

