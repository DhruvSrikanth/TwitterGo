// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import (
	"proj1/semaphore"
)

type RWLock struct {
	nReaders     int
	writingState bool
	sem          *semaphore.Semaphore
}

// Create a new read-write lock
func NewRWLock() *RWLock {
	// Hardcoded limit to 32 readers
	readerCapacity := 32
	nReaders := 0
	writingState := false
	sem := semaphore.NewSemaphore(readerCapacity)
	return &RWLock{nReaders, writingState, sem}
}

// Lock locks rw for writing. If the lock is already locked for reading or writing, Lock blocks until the lock is available.
func (lock *RWLock) Lock() {
	// Lock the semaphore
	lock.sem.Down()
	// Wait until there are no readers
	for lock.writingState {
		lock.sem.Up()
		lock.sem.Down()
	}
	// Someone is writing, wait until they are done
	lock.writingState = true
	// Unlock the semaphore
	lock.sem.Up()
}

// Unlock unlocks rw for writing.
func (lock *RWLock) Unlock() {
	// Lock the semaphore
	lock.sem.Down()
	// Someone is done writing
	lock.writingState = false
	// Unlock the semaphore
	lock.sem.Up()
}

// RLock locks rw for reading.
func (lock *RWLock) RLock() {
	// Lock the semaphore
	lock.sem.Down()
	// Increment the number of readers
	lock.nReaders++
	// Unlock the semaphore
	lock.sem.Up()
}

// RUnlock undoes a single RLock call; it does not affect other simultaneous readers.
func (lock *RWLock) RUnlock() {
	// Lock the semaphore
	lock.sem.Down()
	// Decrement the number of readers
	lock.nReaders--
	// Unlock the semaphore
	lock.sem.Up()
}
