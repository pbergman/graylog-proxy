package net

type ConnQueueItem struct {
	status chan struct{}
	tries  int
	error  []error
	data   []byte
	id     []byte
}

// Tries will return a int representing the amount
// of tries when successfully was delivered, stopped
// because of error or reached max retries
func (c ConnQueueItem) Tries() int {
	return c.tries
}

// Wait will block until queue item is processed
func (c ConnQueueItem) Wait() {
	<-c.status
}

// Error will return the errors
func (c ConnQueueItem) Error() []error {
	return c.error
}

// HasError will check if we got a error
func (c ConnQueueItem) HasError() bool {
	return len(c.error) > 0
}
