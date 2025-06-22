package main

import (
	"fmt"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello from Todo App!\n")
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
