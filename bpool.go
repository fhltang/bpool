package bpool

import (
	"sync"
)

// A BPool is a bounded pool.
type BPool struct {
	semaphore chan interface{}

	pool sync.Pool
}

func New(max int64, new func() interface{}) *BPool {
	bpool := BPool{
		semaphore: make(chan interface{}, max),
		pool:      sync.Pool{New: new},
	}
	return &bpool
}

// Get fetches an item from the pool, blocking if necessary.
//
// If it does not exist, it will call New to create a new one.
func (this *BPool) Get() interface{} {
	this.semaphore <- nil
	return this.pool.Get()
}

// Put returns an item to the pool.
//
// May wake goroutine blocked on a Get call.
func (this *BPool) Put(item interface{}) {
	this.pool.Put(item)
	_ = <-this.semaphore
}
