package bpool_test

import (
	"github.com/fhltang/bpool"
	"sync"
	"testing"
)

func NewString() interface{} {
	return ""
}

func TestNew(t *testing.T) {
	bpool.New(1, NewString)
}

// Test that Get() and Put() do the right thing.
func TestAcquireRelease(t *testing.T) {
	pool := bpool.New(1, NewString)

	s := pool.Get()
	pool.Put(s)
}

// Test that Get() can be called after Put().
func TestAcquireRelease2(t *testing.T) {
	pool := bpool.New(1, NewString)

	s := pool.Get()
	pool.Put(s)
	s = pool.Get()
	pool.Put(s)
}

// Test that the number of Acquired() buffers is bounded.
func TestBounded(t *testing.T) {
	pool := bpool.New(1, NewString)

	// mutex protects released1 which is used by thread 1 to
	// indicate that is has released its string.
	mutex := sync.Mutex{}
	var released1 bool = false

	// Used by thread 1 to indicate that it has acquired a buffer.
	acquired1 := sync.WaitGroup{}
	acquired1.Add(1)

	// Used to allow thread 1 to continue executing and release its item.
	cont1 := sync.WaitGroup{}
	cont1.Add(1)

	// Used to allow thread 2 to indicate that it has released its item.
	released2 := sync.WaitGroup{}
	released2.Add(1)

	thread1 := func() {
		s := pool.Get()
		acquired1.Done()

		cont1.Wait()

		mutex.Lock()
		defer mutex.Unlock()
		pool.Put(s)
		released1 = true
	}

	thread2 := func() {
		s := pool.Get()

		mutex.Lock()
		if !released1 {
			t.Error("Thread 2 acquired a buffer before thread 1 released its buffer.")
		}
		mutex.Unlock()

		pool.Put(s)
		released2.Done()
	}

	go thread1()

	acquired1.Wait()
	go thread2()

	cont1.Done()
	released2.Wait()
}
