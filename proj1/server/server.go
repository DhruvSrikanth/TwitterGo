package server

import (
	"encoding/json"
	"proj1/feed"
	"proj1/queue"
	"sync"
	"sync/atomic"
)

type Config struct {
	Encoder *json.Encoder // Represents the buffer to encode Responses
	Decoder *json.Decoder // Represents the buffer to decode Requests
	Mode    string        // Represents whether the server should execute
	// sequentially or in parallel
	// If Mode == "s"  then run the sequential version
	// If Mode == "p"  then run the parallel version
	// These are the only values for Version
	ConsumersCount int // Represents the number of consumers to spawn
}

type SharedContext struct {
	mutex       *sync.Mutex          // Mutex to use for locking
	cond        *sync.Cond           // Condition variable to use for waiting
	group       *sync.WaitGroup      // Wait group to use for waiting for consumers
	done        bool                 // Flag to indicate if the producer has seen the DONE command
	feed        *feed.Feed           // The twitter feed
	queue       *queue.LockFreeQueue // The queue of requests
	queuedTasks int64                // The number of tasks currently queued by the producer
}

// Run starts up the twitter server based on the configuration
// information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
	// Get the twitter feed
	feed := feed.NewFeed()
	if config.Mode == "s" {
		// Run the sequential version
		sequentialServer(config, feed)
	} else if config.Mode == "p" {
		q := queue.NewLockFreeQueue()
		// Run the parallel version
		parallelServer(config, feed, q)
	}
}

// sequentialServer runs the server in sequential mode
func sequentialServer(config Config, feed feed.Feed) {
	// Loop until we get a DONE command
	for {
		var message map[string]interface{}

		// Decode the request
		err := config.Decoder.Decode(&message)
		if err != nil {
			return
		} else {
			// Exit after seeing the DONE command
			if message["command"] == "DONE" {
				break
			}

			// Wrap the request as a task
			request := queue.Request{Message: message}
			// Process the request
			processRequest(config, feed, request)
		}
	}
}

// parallelServer runs the server in parallel mode
func parallelServer(config Config, feed feed.Feed, q *queue.LockFreeQueue) {
	// Shared context
	group := sync.WaitGroup{}
	mutex := sync.Mutex{}
	cond := sync.NewCond(&mutex)
	context := SharedContext{
		mutex: &mutex,
		cond:  cond,
		group: &group,
		done:  false,
		feed:  &feed,
		queue: q,
	}

	// Spawn the consumers
	for i := 0; i < config.ConsumersCount; i++ {
		context.group.Add(1)
		go consumer(config, &context, i)
	}

	// producer to add requests to the queue
	producer(config, &context)
	context.group.Wait()

}

// consumer processes requests from the queue
func consumer(config Config, context *SharedContext, i int) {
	// Loop until the DONE command is received
	for {
		context.mutex.Lock()
		// If the queue is empty but the producer has not seen the done command, wait
		if context.queuedTasks == 0 && !context.done {
			context.cond.Wait()
		} else if context.queuedTasks == 0 && context.done {
			// If the queue is empty and the producer has seen the done command, exit
			context.mutex.Unlock()
			context.group.Done()
			return
		}

		// Get the next request from the queue
		request := context.queue.Dequeue()
		context.mutex.Unlock()

		// Process the request
		processRequest(config, *context.feed, *request)
		atomic.AddInt64(&context.queuedTasks, int64(-1))

		// If the producer has seen the done command and the queue is empty, exit
		if context.done && context.queuedTasks <= 0 {
			context.group.Done()
			return
		}
	}
}

// producer add requests to the queue
func producer(config Config, context *SharedContext) {
	// Loop until context.done is true
	for {
		// Message to decode into
		var message map[string]interface{}
		// Decode the request
		err := config.Decoder.Decode(&message)
		if err != nil {
			return
		} else {
			// If the command is DONE, set context.done to true
			// And notify the consumers
			// Add return for the producer
			if message["command"] == "DONE" {
				context.done = true
				// Notify all consumers
				context.cond.Broadcast()
				return
			} else {
				// Wrap the request as a task
				request := queue.Request{Message: message}
				// Add the request to the queue
				context.queue.Enqueue(&request)
				// Increment the number of queued tasks
				atomic.AddInt64(&context.queuedTasks, int64(1))
				// Notify 1 consumer if there are any waiting
				context.cond.Signal()
			}
		}
	}
}

// processRequest processes a single request
func processRequest(config Config, feed feed.Feed, request queue.Request) {
	// DONE is included but is checked in a different way
	acceptedCommands := []string{"ADD", "REMOVE", "CONTAINS", "FEED"}
	// Check if it is a valid command
	if request.Message["command"] == nil {
		return
	} else {
		command := request.Message["command"].(string)
		if !contains(acceptedCommands, command) && command != "DONE" {
			// This is an invalid command
			return
		} else if !contains(acceptedCommands, command) && command == "DONE" {
			// Return since we recieved the DONE command
			return
		} else {
			// Get the response as a message
			var response queue.Request
			// This is needed for initialization
			response.Message = make(map[string]interface{})
			response.Message["id"] = request.Message["id"].(float64)

			var success bool

			// Process the request
			switch command {
			case "ADD":
				// Add the post to the feed
				feed.Add(request.Message["body"].(string), request.Message["timestamp"].(float64))
				success = true
			case "REMOVE":
				// Remove the post from the feed and check the success
				success = feed.Remove(request.Message["timestamp"].(float64))
			case "CONTAINS":
				// Check if the post is in the feed
				success = feed.Contains(request.Message["timestamp"].(float64))
			case "FEED":
				// Get the entire feed
				response.Message["feed"] = feed.Show()
			}

			if command != "FEED" {
				response.Message["success"] = success
			}

			// Encode the response
			err := config.Encoder.Encode(&response.Message)
			if err != nil {
				return
			}
		}
	}
}

// contains checks if a string is in a slice of strings
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
