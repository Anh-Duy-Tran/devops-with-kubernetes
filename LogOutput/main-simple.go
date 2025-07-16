package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	pingPongServiceURL = "http://pingpong-service/pingpongcount"
)

var (
	startTime    time.Time
	randomString string
)

func init() {
	startTime = time.Now()
	randomString = uuid.New().String()
}

func main() {
	log.Println("Simple LogOutput starting on port 8080...")

	// Setup HTTP server
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/status", statusHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getPingPongCount() int {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(pingPongServiceURL)
	if err != nil {
		log.Printf("Error calling PingPong service: %v", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("PingPong service returned status: %d", resp.StatusCode)
		return 0
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading PingPong response: %v", err)
		return 0
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		log.Printf("Error parsing PingPong count: %v", err)
		return 0
	}

	return count
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now()
	timestamp := currentTime.Format(time.RFC3339)
	
	// Get current ping-pong count
	pingPongCount := getPingPongCount()

	// Return the required format
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%s: %s.\nPing / Pongs: %d", timestamp, randomString, pingPongCount)
	
	log.Printf("Served status - Ping/Pongs: %d", pingPongCount)
} 