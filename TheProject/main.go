package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Configuration variables - loaded from environment
var (
	imageDirectory string
	imageFileName  string
	imageURL       string
	cacheDuration  time.Duration
	todoBackendURL string
)

func init() {
	// Load configuration from environment variables with defaults
	imageDirectory = getEnvOrDefault("IMAGE_DIRECTORY", "./images")
	imageFileName = getEnvOrDefault("IMAGE_FILENAME", "current.jpg")
	imageURL = getEnvOrDefault("IMAGE_URL", "https://picsum.photos/1200")
	todoBackendURL = getEnvOrDefault("TODO_BACKEND_URL", "http://todo-backend-service:3001")
	
	// Parse cache duration from environment (in minutes)
	cacheDurationMinutes := getEnvOrDefault("CACHE_DURATION_MINUTES", "10")
	minutes, err := strconv.Atoi(cacheDurationMinutes)
	if err != nil {
		fmt.Printf("Invalid CACHE_DURATION_MINUTES: %s, using default 10 minutes\n", cacheDurationMinutes)
		minutes = 10
	}
	cacheDuration = time.Duration(minutes) * time.Minute

	// Create image directory if it doesn't exist
	if err := os.MkdirAll(imageDirectory, 0755); err != nil {
		fmt.Printf("Error creating image directory: %s\n", err)
		os.Exit(1)
	}

	// Parse the HTML template at startup
	var templateErr error
	indexTemplate, templateErr = template.ParseFiles("index.html")
	if templateErr != nil {
		fmt.Printf("Error loading template: %s\n", templateErr)
		os.Exit(1)
	}

	// Print configuration on startup
	fmt.Println("Configuration loaded:")
	fmt.Printf("  Image Directory: %s\n", imageDirectory)
	fmt.Printf("  Image Filename: %s\n", imageFileName)
	fmt.Printf("  Image URL: %s\n", imageURL)
	fmt.Printf("  Cache Duration: %v\n", cacheDuration)
	fmt.Printf("  Todo Backend URL: %s\n", todoBackendURL)

	// Start image refresh goroutine
	go imageRefreshWorker()
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Todo represents a single todo item
type Todo struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Created  string `json:"created"`
	Priority string `json:"priority"`
}

// PageData holds the data to be passed to the HTML template
type PageData struct {
	UserAgent string
	Method    string
	URL       string
	ImagePath string
	ImageAge  string
	Todos     []Todo
}

// Global variable to hold the parsed template
var indexTemplate *template.Template

func imageRefreshWorker() {
	// Check if we need to fetch a new image on startup
	refreshImageIfNeeded()

	// Set up a ticker to check every minute if we need a new image
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		refreshImageIfNeeded()
	}
}

func refreshImageIfNeeded() {
	imagePath := filepath.Join(imageDirectory, imageFileName)

	// Check if image exists and its age
	info, err := os.Stat(imagePath)
	if err != nil {
		// Image doesn't exist, fetch it
		fmt.Println("No cached image found, fetching new image...")
		fetchAndSaveImage()
		return
	}

	age := time.Since(info.ModTime())
	if age > cacheDuration {
		fmt.Printf("Image is %v old, fetching new image...\n", age)
		fetchAndSaveImage()
	}
}

func fetchAndSaveImage() {
	imagePath := filepath.Join(imageDirectory, imageFileName)

	// Fetch image from configured URL
	resp, err := http.Get(imageURL)
	if err != nil {
		fmt.Printf("Error fetching image: %s\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received status code %d when fetching image\n", resp.StatusCode)
		return
	}

	// Create/overwrite the image file
	file, err := os.Create(imagePath)
	if err != nil {
		fmt.Printf("Error creating image file: %s\n", err)
		return
	}
	defer file.Close()

	// Copy image data to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("Error saving image: %s\n", err)
		return
	}

	fmt.Printf("Successfully cached new image to %s\n", imagePath)
}

func getImageAge() string {
	imagePath := filepath.Join(imageDirectory, imageFileName)
	info, err := os.Stat(imagePath)
	if err != nil {
		return "Unknown"
	}

	age := time.Since(info.ModTime())
	if age < time.Minute {
		return fmt.Sprintf("%.0f seconds ago", age.Seconds())
	} else if age < time.Hour {
		return fmt.Sprintf("%.0f minutes ago", age.Minutes())
	} else {
		return fmt.Sprintf("%.1f hours ago", age.Hours())
	}
}

func fetchTodosFromBackend() ([]Todo, error) {
	resp, err := http.Get(todoBackendURL + "/todos")
	if err != nil {
		fmt.Printf("Error fetching todos: %s\n", err)
		return getHardcodedTodos(), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received status code %d when fetching todos\n", resp.StatusCode)
		return getHardcodedTodos(), fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	var todos []Todo
	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		fmt.Printf("Error decoding todos JSON: %s\n", err)
		return getHardcodedTodos(), err
	}

	return todos, nil
}

func createTodoInBackend(text, priority string) error {
	payload := map[string]string{
		"text":     text,
		"priority": priority,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(todoBackendURL+"/todos", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	return nil
}

func getHardcodedTodos() []Todo {
	return []Todo{
		{
			ID:       1,
			Text:     "Set up Kubernetes cluster with persistent volumes",
			Created:  "2 hours ago",
			Priority: "high",
		},
		{
			ID:       2,
			Text:     "Implement image caching functionality",
			Created:  "1 hour ago",
			Priority: "medium",
		},
		{
			ID:       3,
			Text:     "Add todo list functionality to the app",
			Created:  "30 minutes ago",
			Priority: "high",
		},
		{
			ID:       4,
			Text:     "Write comprehensive documentation",
			Created:  "15 minutes ago",
			Priority: "low",
		},
		{
			ID:       5,
			Text:     "Test container restart persistence",
			Created:  "10 minutes ago",
			Priority: "medium",
		},
		{
			ID:       6,
			Text:     "Deploy to production environment",
			Created:  "5 minutes ago",
			Priority: "low",
		},
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	userAgent := req.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "Unknown"
	}

	// Fetch todos from backend service
	todos, err := fetchTodosFromBackend()
	if err != nil {
		fmt.Printf("Using hardcoded todos due to backend error: %s\n", err)
	}

	// Prepare data for template
	data := PageData{
		UserAgent: userAgent,
		Method:    req.Method,
		URL:       req.URL.String(),
		ImagePath: "/image",
		ImageAge:  getImageAge(),
		Todos:     todos,
	}

	// Execute template with data
	err = indexTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		fmt.Printf("Template execution error: %s\n", err)
	}
}

func serveImage(w http.ResponseWriter, req *http.Request) {
	imagePath := filepath.Join(imageDirectory, imageFileName)

	// Check if image exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Serve the image
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "max-age=60") // Cache for 1 minute in browser
	http.ServeFile(w, req, imagePath)
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func healthCheck(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

// Handle todo creation
func createTodo(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := req.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	text := req.FormValue("text")
	priority := req.FormValue("priority")

	// Validate input
	if text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	if len(text) > 140 {
		http.Error(w, "Text must be 140 characters or less", http.StatusBadRequest)
		return
	}

	if priority == "" {
		priority = "medium"
	}

	// Create todo in backend
	if err := createTodoInBackend(text, priority); err != nil {
		fmt.Printf("Error creating todo in backend: %s\n", err)
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// Redirect back to main page
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

// Shutdown endpoint for testing container restarts
func shutdown(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Shutting down server for testing...\n")
	fmt.Println("Shutdown endpoint called - exiting for testing purposes")
	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/image", serveImage)
	http.HandleFunc("/create-todo", createTodo)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/shutdown", shutdown) // For testing container restarts

	// Retrieve port from environment variable
	port := getEnvOrDefault("PORT", "8080")

	// Start HTTP server
	fmt.Printf("Server started on port %s\n", port)
	fmt.Printf("Image cache directory: %s\n", imageDirectory)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
