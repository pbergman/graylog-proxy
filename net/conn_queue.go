package net

import (
	"crypto/sha1"
	"sync"
	"time"
)

type connQueue struct {
	wg    sync.WaitGroup
	queue chan *ConnQueueItem
	tries int
}

func (c connQueue) Wait() {
	c.wg.Wait()
}

func (c *connQueue) newQueueItem(b []byte, id []byte) *ConnQueueItem {
	if nil == id {
		hasher := sha1.New()
		hasher.Write(b)
		now := time.Now().Unix()
		hasher.Write([]byte{
			byte(now),
			byte(now >> 8),
			byte(now >> 16),
			byte(now >> 24),
			byte(now >> 32),
			byte(now >> 40),
			byte(now >> 48),
			byte(now >> 56),
		})
		id = hasher.Sum(nil)
	}
	return &ConnQueueItem{
		status: make(chan struct{}),
		error:  make([]error, 0),
		data:   b,
		id:     id,
	}
}

func (c *connQueue) Push(d []byte, id []byte) *ConnQueueItem {
	data := c.newQueueItem(d, id)
	c.queue <- data
	return data
}

func (c *connQueue) Write(d []byte) (int, error) {
	data := c.newQueueItem(d, nil)
	c.queue <- data
	<-data.status
	return len(d), nil
}
