package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
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
	db    *sql.DB
	mutex sync.Mutex
)

func main() {
	// Initialize database connection
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Setup HTTP server
	http.HandleFunc("/", pingPongHandler)     // Main ping-pong endpoint at root
	http.HandleFunc("/health", healthHandler) // Health check endpoint
	http.HandleFunc("/pingpongcount", pingPongCountHandler)

	log.Println("Starting PingPong server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() (*sql.DB, error) {
	// Get database connection details from environment variables
	dbHost := getEnv("POSTGRES_HOST", "postgres-stset-0.postgres-svc.exercises.svc.cluster.local")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "pingponguser")
	dbPassword := getEnv("POSTGRES_PASSWORD", "pingpongpass")
	dbName := getEnv("POSTGRES_DB", "pingpongdb")

	// Connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Connecting to database at %s:%s", dbHost, dbPort)

	// Retry connection with backoff
	var db *sql.DB
	var err error
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("Successfully connected to database")
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS ping_counter (
		id SERIAL PRIMARY KEY,
		counter_value INTEGER NOT NULL DEFAULT 0
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	// Initialize counter if it doesn't exist
	var count int
	err = db.QueryRow("SELECT counter_value FROM ping_counter WHERE id = 1").Scan(&count)
	if err == sql.ErrNoRows {
		// Insert initial counter value
		_, err = db.Exec("INSERT INTO ping_counter (id, counter_value) VALUES (1, 0)")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize counter: %v", err)
		}
		log.Println("Initialized counter to 0")
	} else if err != nil {
		return nil, fmt.Errorf("failed to query counter: %v", err)
	} else {
		log.Printf("Found existing counter value: %d", count)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getCounter() (int, error) {
	var count int
	err := db.QueryRow("SELECT counter_value FROM ping_counter WHERE id = 1").Scan(&count)
	return count, err
}

func incrementCounter() (int, error) {
	var newCount int
	err := db.QueryRow("UPDATE ping_counter SET counter_value = counter_value + 1 WHERE id = 1 RETURNING counter_value").Scan(&newCount)
	return newCount, err
}

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	// Thread-safe counter increment
	mutex.Lock()
	defer mutex.Unlock()

	currentCount, err := incrementCounter()
	if err != nil {
		log.Printf("Error incrementing counter: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// The counter is incremented, but we want to show the previous value
	displayCount := currentCount - 1

	response := PingPongResponse{
		Message: fmt.Sprintf("pong %d", displayCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Responded with: pong %d", displayCount)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Health check endpoint for Ingress - returns 200 OK
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func pingPongCountHandler(w http.ResponseWriter, r *http.Request) {
	// Return current counter value without incrementing
	mutex.Lock()
	defer mutex.Unlock()

	currentCount, err := getCounter()
	if err != nil {
		log.Printf("Error getting counter: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return as plain text for easy parsing
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(strconv.Itoa(currentCount)))

	log.Printf("Returned count: %d", currentCount)
}
