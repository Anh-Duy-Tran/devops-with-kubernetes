package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// PingPongResponse represents the ping-pong response
type PingPongResponse struct {
	Message string `json:"message"`
}

var (
	counter int
	mutex   sync.Mutex
)

const counterFilePath = "/usr/src/app/files/pingpong_counter.txt"

func main() {
	// Load counter from file if it exists
	loadCounter()

	// Ensure the directory exists
	if err := os.MkdirAll("/usr/src/app/files", 0755); err != nil {
		log.Printf("Failed to create directory: %v", err)
	}

	// Setup HTTP server
	http.HandleFunc("/pingpong", pingPongHandler)
	http.HandleFunc("/", pingPongHandler) // Also handle root path

	log.Println("Starting PingPong server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadCounter() {
	if data, err := os.ReadFile(counterFilePath); err == nil {
		if count, err := strconv.Atoi(string(data)); err == nil {
			counter = count
			log.Printf("Loaded counter from file: %d", counter)
		}
	}
}

func saveCounter(count int) {
	if err := os.WriteFile(counterFilePath, []byte(strconv.Itoa(count)), 0644); err != nil {
		log.Printf("Error saving counter to file: %v", err)
	}
}

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	// Thread-safe counter increment
	mutex.Lock()
	currentCount := counter
	counter++
	saveCounter(currentCount) // Save the incremented counter (next value to be returned)
	mutex.Unlock()

	response := PingPongResponse{
		Message: fmt.Sprintf("pong %d", currentCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Responded with: pong %d", currentCount)
}
