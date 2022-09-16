package concurrency

import (
	"github.com/xefino/goutils/collections"
)

// Buffer allows for data to be requested in a quantity that is limited by a single parameter
type Buffer[T any] struct {
	data    *collections.ConcurrentList[*T]
	size    int
	limiter chan interface{}
}

// NewBuffer creates a new Buffer that allows concurrent operation on size elements of data
func NewBuffer[T any](size int) *Buffer[T] {
	return &Buffer[T]{
		data:    collections.NewConcurrentList[*T](),
		size:    size,
		limiter: make(chan interface{}, size),
	}
}

// Size returns the number of items of data that may be requested concurrently
func (buf *Buffer[T]) Size() int {
	return buf.size
}

// Load adds data to the buffer. This function may be called concurrently
func (buf *Buffer[T]) Load(data ...*T) {
	buf.data.Append(data...)
}

// Get retrieves an item from the buffer. This function will block if the number of concurrent
// operations is at the value limited by size. Otherwise, an item will be removed from the data
// channel, if it exists, and returned. If data was in the channel then true will be returned as
// well. Otherwise, nil and false will be returned. The Release function should be called once
// operation on this data has ceased
func (buf *Buffer[T]) Get() (*T, bool) {

	// Increase our concurrency count; this will block if the limit has been reached already
	buf.limiter <- new(interface{})

	// Request the data; if there is any in the list then return it and true
	if buf.data.Length() >= 1 {
		return buf.data.PopFront(1)[0], true
	}

	// There was no data in the channel so return nil, false
	<-buf.limiter
	return nil, false
}

// Release informs the buffer that work on one of the data items has completed and that another
// operation may begin.
func (buf *Buffer[T]) Release() {
	<-buf.limiter
}
