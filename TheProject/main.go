package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

// PageData holds the data to be passed to the HTML template
type PageData struct {
	UserAgent string
	Method    string
	URL       string
}

// Global variable to hold the parsed template
var indexTemplate *template.Template

func init() {
	// Parse the HTML template at startup
	var err error
	indexTemplate, err = template.ParseFiles("index.html")
	if err != nil {
		fmt.Printf("Error loading template: %s\n", err)
		os.Exit(1)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	userAgent := req.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "Unknown"
	}

	// Prepare data for template
	data := PageData{
		UserAgent: userAgent,
		Method:    req.Method,
		URL:       req.URL.String(),
	}

	// Execute template with data
	err := indexTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		fmt.Printf("Template execution error: %s\n", err)
	}
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

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/health", healthCheck)

	// Retrieve port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	// Start HTTP server
	fmt.Printf("Server started in port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
