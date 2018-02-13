package net

import (
    "net/http"
    "sync"
    "bytes"
    "fmt"

    "github.com/pbergman/logger"
 )

type HttpConnPool struct {
    connQueue
    clients     []*http.Client
    host        *GraylogHost
    lock        sync.Mutex
    pool        *sync.Pool
    logger      logger.LoggerInterface
}

func (p HttpConnPool) Close() {
    close(p.queue)
}

func (p HttpConnPool) Start(workers int) {
    p.clients = make([]*http.Client, workers)
    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        p.clients[i] = &http.Client{}
        go p.process(p.clients[i], &p.wg)
    }
}

func (p *HttpConnPool) post(item *ConnQueueItem, conn *http.Client) error {
    buf := p.pool.Get().(*bytes.Buffer)
    defer p.pool.Put(buf)
    defer buf.Reset()
    if _, err := buf.Write(item.data); err != nil {
        return err
    }
    p.logger.Info(fmt.Sprintf("[%X] [POST] %d bytes to '%s'", item.id, buf.Len(), p.host.String()))
    request, err := http.NewRequest("POST", p.host.String(), buf)
    if err != nil {
        return err
    }
    response, err := conn.Do(request)
    if err != nil {
        return err
    }
    p.logger.Info(fmt.Sprintf("[%X] [POST] %s", item.id, response.Status))
    return nil

}

func (p *HttpConnPool) process(conn *http.Client, wg *sync.WaitGroup) {
    defer wg.Done()
    for item := range p.queue {
        if err := p.post(item, conn); err != nil {
            item.tries++
            item.error = append(item.error, err)
            p.logger.Debug(fmt.Sprintf("[%X] %#v", item.id, err))
            p.logger.Error(fmt.Sprintf("[%X] %s", item.id, err.Error()))
            if item.tries < p.tries {
                p.queue <- item
            } else {
                p.logger.Alert(fmt.Sprintf("[%X] discarded message after %d retires", item.id, item.tries))
                close(item.status)
            }

        } else {
            close(item.status)
        }
    }
}

func NewHttpConnPool(tries int, host *GraylogHost, logger logger.LoggerInterface) (ConnPoolInterface, error) {
    return &HttpConnPool{
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
    }, nil
}
