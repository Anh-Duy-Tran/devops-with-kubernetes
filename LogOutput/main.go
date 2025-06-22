package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Status represents the application status
type Status struct {
	Timestamp string `json:"timestamp"`
	String    string `json:"string"`
	PingPongs int    `json:"ping_pongs,omitempty"`
}

const (
	logFilePath         = "/usr/src/app/files/output.log"
	pingPongCounterPath = "/usr/src/app/files/pingpong_counter.txt"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: log-output [writer|reader]")
	}

	mode := os.Args[1]

	switch mode {
	case "writer":
		runWriter()
	case "reader":
		runReader()
	default:
		log.Fatal("Invalid mode. Use 'writer' or 'reader'")
	}
}

func runWriter() {
	// Generate random string on startup
	id := uuid.New()
	randomString := id.String()

	log.Printf("Writer mode started with string: %s", randomString)

	// Ensure the directory exists
	if err := os.MkdirAll("/usr/src/app/files", 0755); err != nil {
		log.Fatal("Failed to create directory:", err)
	}

	// Write log entries every 5 seconds
	for {
		currentTime := time.Now()
		timestamp := currentTime.Format(time.RFC3339)

		// Get ping-pong count
		pingPongCount := readPingPongCount()

		logEntry := fmt.Sprintf("%s: %s.\nPing / Pongs: %d\n", timestamp, randomString, pingPongCount)

		// Append to file
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("Error opening file: %v", err)
		} else {
			file.WriteString(logEntry)
			file.Close()
		}

		// Also print to stdout for logging
		fmt.Print(logEntry)

		time.Sleep(5 * time.Second)
	}
}

func readPingPongCount() int {
	if data, err := os.ReadFile(pingPongCounterPath); err == nil {
		if count, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			return count
		}
	}
	return 0
}

func runReader() {
	log.Println("Reader mode started, serving on port 8080...")

	// Setup HTTP server
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/status", statusHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	// Read the log file
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		// If file doesn't exist, return waiting message
		response := Status{
			Timestamp: time.Now().Format(time.RFC3339),
			String:    "Waiting for log data...",
			PingPongs: readPingPongCount(),
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s: %s.\nPing / Pongs: %d", response.Timestamp, response.String, response.PingPongs)
		return
	}

	// Return the file content as-is (it's already in the required format)
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}
