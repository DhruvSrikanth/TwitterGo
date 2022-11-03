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
		tailNode := (*node)(atomic.LoadPointer(&queue.tail))
		nextNode := (*node)(atomic.LoadPointer(&tailNode.next))
		// If the tail is the same as the current tail
		if tailNode == (*node)(atomic.LoadPointer(&queue.tail)) {
			// If the tail is the last node in the queue
			if nextNode == nil {
				// Try to add the node to the queue
				if atomic.CompareAndSwapPointer(&tailNode.next, unsafe.Pointer(nextNode), unsafe.Pointer(node_)) {
					atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(tailNode), unsafe.Pointer(node_))
					return
				}
			} else {
				// If the tail is not the last node in the queue
				atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(tailNode), unsafe.Pointer(nextNode))
			}
		}
	}
}

// Reference: https://www.sobyte.net/post/2021-07/implementing-lock-free-queues-with-go/
// Dequeue removes a Request from the queue
func (queue *LockFreeQueue) Dequeue() *Request {
	for {
		// Traverse the queue
		headNode := (*node)(atomic.LoadPointer(&queue.head))
		tailNode := (*node)(atomic.LoadPointer(&queue.tail))
		nextNode := (*node)(atomic.LoadPointer(&headNode.next))
		// If the head is the same as the current head
		if headNode == (*node)(atomic.LoadPointer(&queue.head)) {
			// If the head is the same as the tail
			if headNode == tailNode {
				// If the queue is empty
				if nextNode == nil {
					return &Request{Message: nil}
				}
				atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(tailNode), unsafe.Pointer(nextNode))
			} else {
				// If the head is not the same as the tail
				request := nextNode.value
				if atomic.CompareAndSwapPointer(&queue.head, unsafe.Pointer(headNode), unsafe.Pointer(nextNode)) {
					return &request
				}
			}
		}

	}
}
