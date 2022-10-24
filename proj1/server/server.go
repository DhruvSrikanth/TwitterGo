package server

import "encoding/json"

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

//Run starts up the twitter server based on the configuration
//information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
	panic("TODO")

}
