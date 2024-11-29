package main

import (
	"encoding/json"
	"log"
	"net/http"

	tollbooth "github.com/didip/tollbooth/v7"
	//"github.com/didip/tollbooth/v8/limiter"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Successful",
		Body:   "Hi! You're reached the test API For TollBooth. How can I help you?",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

func main() {
	message := Message{
		Status: "Request Failed",
		Body:   "API is at capacity, try again later.",
	}

	jsonMessage, _ := json.Marshal(message)
	tollboothLimiter := tollbooth.NewLimiter(1, nil)
	tollboothLimiter.SetMessageContentType("application/json")
	tollboothLimiter.SetMessage(string(jsonMessage))
	// Just Useful for toolbooth v8
	// New in version >= 8, you must explicitly define how to pick the IP address.
	// tollboothLimiter.SetIPLookup(limiter.IPLookup{
	// 	Name:           "RemoteAddr",
	// 	IndexFromRight: 0,
	// })
	http.Handle("/ping", tollbooth.LimitFuncHandler(tollboothLimiter, endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("there was an error while starting the server on Port 8080")
	}
}
