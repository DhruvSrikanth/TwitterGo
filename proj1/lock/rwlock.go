// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import (
	"proj1/semaphore"
)

type RWLock struct {
	readerCapacity int
	writerSem      *semaphore.Semaphore
	readerSem      *semaphore.Semaphore
}

// Create a new read-write lock
func NewRWLock() *RWLock {
	// Hardcoded limit to 32 readers
	readerCapacity := 32
	writerCapacity := 1
	writerSem := semaphore.NewSemaphore(writerCapacity)
	readerSem := semaphore.NewSemaphore(readerCapacity)
	return &RWLock{readerCapacity, writerSem, readerSem}
}

// Lock locks rw for writing. If the lock is already locked for reading or writing, Lock blocks until the lock is available.
func (lock *RWLock) Lock() {
	// Lock the writer semaphore
	lock.writerSem.Down()
	// Lock the reader semaphore
	for i := 0; i < lock.readerCapacity; i++ {
		lock.readerSem.Down()
	}
}

// Unlock unlocks rw for writing.
func (lock *RWLock) Unlock() {
	// Unlock the writer semaphore
	lock.writerSem.Up()
	// Unlock the reader semaphore
	for i := 0; i < lock.readerCapacity; i++ {
		lock.readerSem.Up()
	}
}

// RLock locks rw for reading.
func (lock *RWLock) RLock() {
	// Lock reader the semaphore
	lock.readerSem.Down()
}

// RUnlock undoes a single RLock call; it does not affect other simultaneous readers.
func (lock *RWLock) RUnlock() {
	// Unlock the reader semaphore
	lock.readerSem.Up()
}
