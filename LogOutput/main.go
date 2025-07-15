package main

import (
	"fmt"
	"io"
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
	logFilePath        = "/tmp/output.log"
	pingPongServiceURL = "http://pingpong-service:8080/pingpongcount"
	configFilePath     = "/config/information.txt"
)

func readConfigFile() string {
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return "config file not found"
	}
	return strings.TrimSpace(string(content))
}

func getEnvMessage() string {
	message := os.Getenv("MESSAGE")
	if message == "" {
		return "env variable not set"
	}
	return message
}

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

	// Write log entries every 5 seconds
	for {
		currentTime := time.Now()
		timestamp := currentTime.Format(time.RFC3339)

		// Get ping-pong count via HTTP
		pingPongCount := getPingPongCountHTTP()

		// Read config file content
		fileContent := readConfigFile()

		// Read environment variable
		envMessage := getEnvMessage()

		logEntry := fmt.Sprintf("file content: %s\nenv variable: MESSAGE=%s\n%s: %s.\nPing / Pongs: %d\n", 
			fileContent, envMessage, timestamp, randomString, pingPongCount)

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

func getPingPongCountHTTP() int {
	client := &http.Client{
		Timeout: 5 * time.Second,
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
		// If file doesn't exist, return waiting message with config info
		pingPongCount := getPingPongCountHTTP()
		fileContent := readConfigFile()
		envMessage := getEnvMessage()
		
		response := Status{
			Timestamp: time.Now().Format(time.RFC3339),
			String:    "Waiting for log data...",
			PingPongs: pingPongCount,
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "file content: %s\nenv variable: MESSAGE=%s\n%s: %s.\nPing / Pongs: %d", 
			fileContent, envMessage, response.Timestamp, response.String, response.PingPongs)
		return
	}

	// Return the file content as-is (it's already in the required format)
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}
