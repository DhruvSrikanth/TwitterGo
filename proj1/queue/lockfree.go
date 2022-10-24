package queue

type Request struct {
}

// LockfreeQueue represents a FIFO structure with operations to enqueue
// and dequeue tasks represented as Request
type LockFreeQueue struct {
}

// NewQueue creates and initializes a LockFreeQueue
func NewLockFreeQueue() *LockFreeQueue {

}

// Enqueue adds a series of Request to the queue
func (queue *LockFreeQueue) Enqueue(task *Request) {

}

// Dequeue removes a Request from the queue
func (queue *LockFreeQueue) Dequeue() *Request {
	return nil
}
