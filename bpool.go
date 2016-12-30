package bpool

import (
	"sync"
)

// A BPool is a bounded pool.
//
// Compared to sync.Pool, BPool
//   - manages at most max items
//   - never automatically removes items from the pool.
//
// For correctness, we assume that all managed items were acquired via
// Get().  That is, any item inserted via Put() must originally have
// been acquired through Get().
//
// Behaviour:
//   - new is called at most max times
//   - new is called when needed: the pool is lazy
//   - Get() blocks until an item is available.
//
type BPool struct {
	max int
	new func() interface{}

	// Conditional variable protecting the following variables.
	cond *sync.Cond

	// Capacity.  This is the total number of items managed by
	// this pool, which is equivalent to the number of times we
	// have invoked new().  We assume cap <= max at all times.
	//
	// cap is monotonic; it never decreases.
	cap int

	// Free list.  Items that are managed by this pool that are
	// available.  If freeList is empty, provided cap < max, we
	// may call new to return a new item.
	freeList []interface{}
}

func New(max int, new func() interface{}) *BPool {
	bpool := BPool{
		max: max,
		new: new,

		cond:     sync.NewCond(&sync.Mutex{}),
		cap:      0,
		freeList: make([]interface{}, 0, max),
	}
	return &bpool
}

// Get fetches an item from the pool, blocking if necessary.
//
// If it does not exist, it will call new to create a new one.
func (this *BPool) Get() interface{} {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()

	for len(this.freeList) == 0 && this.cap == this.max {
		this.cond.Wait()
	}
	// We know that either
	//
	//     len(this.freeList) > 0
	//
	// or
	//
	//     this.cap < this.max

	if len(this.freeList) == 0 {
		this.cap++
		return this.new()
	}
	// Necessarily len(this.freeList) > 0 .

	item := this.freeList[len(this.freeList)-1]
	this.freeList = this.freeList[:len(this.freeList)-1]
	return item
}

// Put returns an item to the pool.
//
// May wake goroutine blocked on a Get call.
func (this *BPool) Put(item interface{}) {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()

	this.freeList = append(this.freeList, item)
	this.cond.Signal()
}
