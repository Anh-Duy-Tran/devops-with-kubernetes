package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Status represents the application status
type Status struct {
	Timestamp string `json:"timestamp"`
	String    string `json:"string"`
}

var appStatus Status

func main() {
	// Generate random string on startup
	id := uuid.New()
	appStatus.String = id.String()

	// Start the log output goroutine
	go func() {
		for {
			currentTime := time.Now()
			appStatus.Timestamp = currentTime.Format(time.RFC3339)
			fmt.Printf("[%s]: %s\n", appStatus.Timestamp, appStatus.String)
			time.Sleep(5 * time.Second)
		}
	}()

	// Setup HTTP server
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/status", statusHandler)

	log.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	// Update timestamp for each request
	appStatus.Timestamp = time.Now().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appStatus)
}
