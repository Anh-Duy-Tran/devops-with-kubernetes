package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Status represents the application status
type Status struct {
	Timestamp string `json:"timestamp"`
	String    string `json:"string"`
}

const logFilePath = "/usr/src/app/files/output.log"

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
		logEntry := fmt.Sprintf("[%s]: %s\n", timestamp, randomString)

		// Append to file
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse the last line to get the latest entry
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		response := Status{
			Timestamp: time.Now().Format(time.RFC3339),
			String:    "No log data available",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the last line and parse it
	lastLine := lines[len(lines)-1]

	// Extract timestamp and string from the format "[timestamp]: string"
	parts := strings.SplitN(lastLine, "]: ", 2)
	var timestamp, randomString string

	if len(parts) == 2 {
		timestamp = strings.TrimPrefix(parts[0], "[")
		randomString = parts[1]
	} else {
		timestamp = time.Now().Format(time.RFC3339)
		randomString = lastLine
	}

	response := Status{
		Timestamp: timestamp,
		String:    randomString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
