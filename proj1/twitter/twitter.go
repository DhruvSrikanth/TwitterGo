package main

import (
	"encoding/json"
	parser "flag"
	"fmt"
	"os"
	"proj1/server"
	"strconv"
)

func Usage() {
	fmt.Println("Usage: twitter <number of consumers> \n <number of consumers> = the number of goroutines (i.e., consumers) to be part of the parallel version.")
}

func main() {
	// Create the streaming encoder and decoder
	encoder := json.NewEncoder(os.Stdout)
	decoder := json.NewDecoder(os.Stdin)

	parser.Parse()
	// Get the non flag arguments
	args := parser.Args()

	var mode string
	var config server.Config

	// Error checking was not asked for but it I felt like after the rest of the code was done, it would be nice to include it.
	if len(args) == 0 {
		mode = "s"
		config = server.Config{Encoder: encoder, Decoder: decoder, Mode: mode}
	} else if len(args) == 1 {
		mode = "p"
		nConsumers := args[0]
		numConsumers, err := strconv.Atoi(nConsumers)
		if err != nil {
			fmt.Println("Error: ", err)
			Usage()
			return
		}
		config = server.Config{Encoder: encoder, Decoder: decoder, Mode: mode, ConsumersCount: numConsumers}

	} else {
		Usage()
		return
	}
	// Run the server
	server.Run(config)

}
