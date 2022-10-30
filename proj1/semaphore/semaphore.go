package semaphore

import (
	"sync"
)

type Semaphore struct {
	capacity int
	lock     *sync.Mutex
	cond     *sync.Cond
}

func NewSemaphore(capacity int) *Semaphore {
	lock := sync.Mutex{}
	cond := sync.NewCond(&lock)
	return &Semaphore{capacity, &lock, cond}
}

func (s *Semaphore) Up() {
	// Lock
	s.lock.Lock()
	// Increment capacity
	s.capacity++
	// Wake up a goroutine if it is waiting
	s.cond.Signal()
	// Unlock
	s.lock.Unlock()
}

func (s *Semaphore) Down() {
	// Lock
	s.lock.Lock()

	// We know capacity is non-negative
	// So spin until it is positive
	// Capacity = 0 means that the shared resource is in use to its full capacity
	for s.capacity == 0 {
		// Suspend goroutine and release lock
		s.cond.Wait()
	}

	// Unlock
	s.lock.Unlock()

	// Decrement capacity
	s.capacity--
}
