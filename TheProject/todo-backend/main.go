package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// Todo represents a single todo item
type Todo struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Created  string `json:"created"`
	Priority string `json:"priority"`
}

// CreateTodoRequest represents the request body for creating a new todo
type CreateTodoRequest struct {
	Text     string `json:"text"`
	Priority string `json:"priority,omitempty"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize the database schema
	if err := initSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Seed database with initial data if empty
	if err := seedInitialData(); err != nil {
		log.Printf("Warning: Failed to seed initial data: %v", err)
	}

	// Routes
	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET", "OPTIONS":
			getTodos(w, r)
		case "POST":
			createTodo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/stats", getStats)

	// Get port from environment or use default
	port := getEnvOrDefault("PORT", "3001")

	log.Printf("Todo backend starting on port %s", port)
	log.Printf("Database connection: %s:%s", 
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initDB() (*sql.DB, error) {
	// Get database connection details from environment variables
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("POSTGRES_USER", "todouser")
	dbPassword := getEnvOrDefault("POSTGRES_PASSWORD", "todopass123")
	dbName := getEnvOrDefault("POSTGRES_DB", "tododb")
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")

	// Connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	log.Printf("Connecting to database at %s:%s", dbHost, dbPort)

	// Retry connection with backoff
	var database *sql.DB
	var err error
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		database, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = database.Ping()
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

	return database, nil
}

func initSchema() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		text TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		priority VARCHAR(10) DEFAULT 'medium'
	);
	
	CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at DESC);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	log.Println("Database schema initialized")
	return nil
}

func seedInitialData() error {
	// Check if we already have data
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %v", err)
	}

	if count > 0 {
		log.Printf("Database already has %d todos, skipping seed", count)
		return nil
	}

	// Insert initial todos
	initialTodos := []struct {
		text     string
		priority string
	}{
		{"Set up Kubernetes cluster with persistent volumes", "high"},
		{"Implement image caching functionality", "medium"},
		{"Add todo list functionality to the app", "high"},
		{"Write comprehensive documentation", "low"},
		{"Test container restart persistence", "medium"},
		{"Deploy to production environment", "low"},
	}

	for _, todo := range initialTodos {
		_, err := db.Exec(
			"INSERT INTO todos (text, priority) VALUES ($1, $2)",
			todo.text, todo.priority,
		)
		if err != nil {
			return fmt.Errorf("failed to insert initial todo: %v", err)
		}
	}

	log.Printf("Seeded database with %d initial todos", len(initialTodos))
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func formatCreatedTime(createdAt time.Time) string {
	now := time.Now()
	diff := now.Sub(createdAt)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// CORS middleware
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// GET /todos - Get all todos
func getTodos(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	rows, err := db.Query("SELECT id, text, created_at, priority FROM todos ORDER BY created_at DESC")
	if err != nil {
		log.Printf("Error querying todos: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		var createdAt time.Time
		
		err := rows.Scan(&todo.ID, &todo.Text, &createdAt, &todo.Priority)
		if err != nil {
			log.Printf("Error scanning todo: %v", err)
			continue
		}
		
		todo.Created = formatCreatedTime(createdAt)
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating todos: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)

	log.Printf("GET /todos - Returned %d todos", len(todos))
}

// POST /todos - Create a new todo
func createTodo(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	if len(req.Text) > 140 {
		http.Error(w, "Text must be 140 characters or less", http.StatusBadRequest)
		return
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = "medium"
	}

	// Validate priority
	if req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" {
		req.Priority = "medium"
	}

	// Insert into database
	var newTodo Todo
	var createdAt time.Time
	err := db.QueryRow(
		"INSERT INTO todos (text, priority) VALUES ($1, $2) RETURNING id, text, created_at, priority",
		req.Text, req.Priority,
	).Scan(&newTodo.ID, &newTodo.Text, &createdAt, &newTodo.Priority)

	if err != nil {
		log.Printf("Error creating todo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	newTodo.Created = formatCreatedTime(createdAt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)

	log.Printf("POST /todos - Created todo: %s (ID: %d)", newTodo.Text, newTodo.ID)
}

// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := db.Ping(); err != nil {
		log.Printf("Health check failed - database error: %v", err)
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// Stats endpoint for debugging
func getStats(w http.ResponseWriter, r *http.Request) {
	var totalTodos int
	err := db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&totalTodos)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"total_todos": totalTodos,
		"timestamp":   time.Now().Format(time.RFC3339),
		"database":    "postgres",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)

	log.Printf("GET /stats - Current stats: %d todos", totalTodos)
}

