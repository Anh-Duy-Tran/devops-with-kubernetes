package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
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

// In-memory storage for todos
var (
	todos   []Todo
	todosMu sync.RWMutex
	nextID  = 1
)

func init() {
	// Initialize with some hardcoded todos
	todos = []Todo{
		{
			ID:       nextID,
			Text:     "Set up Kubernetes cluster with persistent volumes",
			Created:  "2 hours ago",
			Priority: "high",
		},
		{
			ID:       nextID + 1,
			Text:     "Implement image caching functionality",
			Created:  "1 hour ago",
			Priority: "medium",
		},
		{
			ID:       nextID + 2,
			Text:     "Add todo list functionality to the app",
			Created:  "30 minutes ago",
			Priority: "high",
		},
		{
			ID:       nextID + 3,
			Text:     "Write comprehensive documentation",
			Created:  "15 minutes ago",
			Priority: "low",
		},
		{
			ID:       nextID + 4,
			Text:     "Test container restart persistence",
			Created:  "10 minutes ago",
			Priority: "medium",
		},
		{
			ID:       nextID + 5,
			Text:     "Deploy to production environment",
			Created:  "5 minutes ago",
			Priority: "low",
		},
	}
	nextID += 6
}

// CORS middleware
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// GET /todos - Fetch all todos
func getTodos(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	todosMu.RLock()
	defer todosMu.RUnlock()

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

	todosMu.Lock()
	defer todosMu.Unlock()

	// Create new todo
	newTodo := Todo{
		ID:       nextID,
		Text:     req.Text,
		Created:  "just now",
		Priority: req.Priority,
	}

	// Add to the beginning of the list (newest first)
	todos = append([]Todo{newTodo}, todos...)
	nextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)

	log.Printf("POST /todos - Created todo: %s (ID: %d)", newTodo.Text, newTodo.ID)
}

// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// Stats endpoint for debugging
func getStats(w http.ResponseWriter, r *http.Request) {
	todosMu.RLock()
	defer todosMu.RUnlock()

	stats := map[string]interface{}{
		"total_todos": len(todos),
		"next_id":     nextID,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)

	log.Printf("GET /stats - Current stats: %d todos, next ID: %d", len(todos), nextID)
}

func main() {
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

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Todo Backend starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /todos  - Fetch all todos")
	log.Printf("  POST /todos  - Create new todo")
	log.Printf("  GET  /health - Health check")
	log.Printf("  GET  /stats  - Service stats")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

