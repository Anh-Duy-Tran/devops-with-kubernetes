package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// PingPongResponse represents the ping-pong response
type PingPongResponse struct {
	Message string `json:"message"`
}

// CountResponse represents just the count
type CountResponse struct {
	Count int `json:"count"`
}

var (
	counter int
	mutex   sync.Mutex
)

func main() {
	// Setup HTTP server
	http.HandleFunc("/pingpong", pingPongHandler)
	http.HandleFunc("/pingpongcount", pingPongCountHandler)
	http.HandleFunc("/", pingPongHandler) // Also handle root path

	log.Println("Starting PingPong server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	// Thread-safe counter increment
	mutex.Lock()
	currentCount := counter
	counter++
	mutex.Unlock()

	response := PingPongResponse{
		Message: fmt.Sprintf("pong %d", currentCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Responded with: pong %d", currentCount)
}

func pingPongCountHandler(w http.ResponseWriter, r *http.Request) {
	// Return current counter value without incrementing
	mutex.Lock()
	currentCount := counter
	mutex.Unlock()

	// Return as plain text for easy parsing
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(strconv.Itoa(currentCount)))

	log.Printf("Returned count: %d", currentCount)
}
