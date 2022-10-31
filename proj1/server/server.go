package server

import (
	"encoding/json"
	"proj1/feed"
	"proj1/queue"
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

// Run starts up the twitter server based on the configuration
// information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
	feed := feed.NewFeed()
	if config.Mode == "s" {
		// Run the sequential version
		seqentualServer(config, feed)
	} else if config.Mode == "p" {
		q := queue.NewLockFreeQueue()
		// Run the parallel version
		parallelServer(config, feed, q)
	}
}

// seqentualServer runs the server in sequential mode
func seqentualServer(config Config, feed feed.Feed) {
	for {
		var message map[string]interface{}

		// Decode the request
		err := config.Decoder.Decode(&message)
		if err != nil {
			break
		} else {
			request := queue.Request{message: message}
			processRequest(config, feed, request)
		}
	}
}

// parallelServer runs the server in parallel mode
func parallelServer(config Config, feed feed.Feed, q *queue.LockFreeQueue) {
}

func processRequest(config Config, feed feed.Feed, request queue.Request) {
	// response
	var response queue.Request

	// Get the request type
	if request.message["command"] == "ADD" {
		feed.Add(request["body"].(string), request["timestamp"].(float64))
	}

}
