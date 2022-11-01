package server

import (
	"encoding/json"
	"proj1/feed"
	"proj1/queue"
	"sync"
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
	mutex *sync.Mutex
	cond  *sync.Cond
	done  bool
}

// Run starts up the twitter server based on the configuration
// information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
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
	for {
		var message map[string]interface{}

		// Decode the request
		err := config.Decoder.Decode(&message)
		if err != nil {
			break
		} else {
			request := queue.Request{Message: message}
			processRequest(config, feed, request)
		}
	}
}

// parallelServer runs the server in parallel mode
func parallelServer(config Config, feed feed.Feed, q *queue.LockFreeQueue) {
	// Shared context
	mutex := sync.Mutex{}
	cond := sync.NewCond(&mutex)
	context := SharedContext{
		mutex: &mutex,
		cond:  cond,
	}

	// Spawn the consumers
	for i := 0; i < config.ConsumersCount; i++ {
		go consumer(config, feed, q, &context)
	}

	// producer
	producer(config, q, &context)

}

// consumer processes requests from the queue
func consumer(config Config, feed feed.Feed, q *queue.LockFreeQueue, context *SharedContext) {
	for {
		if context.done {
			break
		}

		// Wait for a request
		request := q.Dequeue()

		if request == nil {
			context.cond.Wait()
		} else if request.Message["command"] == "DONE" {
			context.mutex.Lock()
			context.done = true
			context.mutex.Unlock()
			context.cond.Broadcast()
		} else {
			// Process the request
			processRequest(config, feed, *request)
		}
	}
}

// producer add requests to the queue
func producer(config Config, q *queue.LockFreeQueue, context *SharedContext) {
	for {
		if context.done {
			break
		}

		var message map[string]interface{}

		// Decode the request
		err := config.Decoder.Decode(&message)

		if err != nil {
			// fmt.Printf("Error: %v while decoding request!\n", err)
		} else {
			request := queue.Request{Message: message}
			// Add the request to the queue
			q.Enqueue(&request)
			// Notify 1 consumer if there are any waiting
			context.cond.Signal()
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
			return
		} else if !contains(acceptedCommands, command) && command == "DONE" {
			return
		} else {
			// response
			var response queue.Request
			// This is needed for initialization
			response.Message = make(map[string]interface{})
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

			// Create the response
			response.Message["success"] = success
			response.Message["id"] = request.Message["id"].(float64)

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
