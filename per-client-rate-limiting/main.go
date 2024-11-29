package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func perclientRateLimiter(next func(writer http.ResponseWriter, response *http.Request)) http.Handler {

	// This should be defined outside of this function
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// This should be passed to the funtion
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ip, _, err := net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(2, 4),
				//lastSeen: time.Now(),
			}
			clients[ip].lastSeen = time.Now()
		}
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			message := Message{
				Status: "Request Failed",
				Body:   "API is at capacity, try again later.",
			}
			writer.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(writer).Encode(&message)
		} else {
			mu.Unlock()
			next(writer, request)
		}
	})

}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Successful",
		Body:   "Hi! You're reached the test API For Client Rate Limiting. How can I help you?",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

func main() {
	http.Handle("/ping", perclientRateLimiter(endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("an error happenned when listening on Port 8080")
	}

}
