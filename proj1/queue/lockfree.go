package queue

import (
	"sync/atomic"
	"unsafe"
)

type Request struct {
	Message map[string]interface{}
}

type node struct {
	value Request
	next  unsafe.Pointer
}

// LockfreeQueue represents a FIFO structure with operations to enqueue
// and dequeue tasks represented as Request
type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

// NewQueue creates and initializes a LockFreeQueue
func NewLockFreeQueue() *LockFreeQueue {
	node_ := unsafe.Pointer(&node{})
	return &LockFreeQueue{head: node_, tail: node_}
}

// Reference: https://www.sobyte.net/post/2021-07/implementing-lock-free-queues-with-go/
// Enqueue adds a series of Request to the queue
func (queue *LockFreeQueue) Enqueue(task *Request) {
	// Node to add to the queue
	node_ := &node{value: *task}
	for {
		// Traverse the queue
		tail := (*node)(atomic.LoadPointer(&queue.tail))
		next := (*node)(atomic.LoadPointer(&tail.next))
		// If the tail is the same as the current tail
		if tail == (*node)(atomic.LoadPointer(&queue.tail)) {
			// If the tail is the last node in the queue
			if next == nil {
				// Try to add the node to the queue
				if atomic.CompareAndSwapPointer(&tail.next, unsafe.Pointer(next), unsafe.Pointer(node_)) {
					atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(tail), unsafe.Pointer(node_))
					return
				}
			} else {
				// If the tail is not the last node in the queue
				atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			}
		}
	}
}

// Reference: https://www.sobyte.net/post/2021-07/implementing-lock-free-queues-with-go/
// Dequeue removes a Request from the queue
func (queue *LockFreeQueue) Dequeue() *Request {
	for {
		// Traverse the queue
		head := (*node)(atomic.LoadPointer(&queue.head))
		tail := (*node)(atomic.LoadPointer(&queue.tail))
		next := (*node)(atomic.LoadPointer(&head.next))
		// If the head is the same as the current head
		if head == (*node)(atomic.LoadPointer(&queue.head)) {
			// If the head is the same as the tail
			if head == tail {
				// If the queue is empty
				if next == nil {
					return &Request{Message: nil}
				}
				atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			} else {
				// If the head is not the same as the tail
				request := next.value
				if atomic.CompareAndSwapPointer(&queue.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
					return &request
				}
			}
		}

	}
}
