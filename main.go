package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	filename := "request_counter.gob"
	rc := NewRequestCounter(filename)

	// Save data periodically
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := rc.SaveToFile(); err != nil {
				log.Fatal("Error saving data to file:", err)
			}
		}
	}()

	// Remove old keys periodically
	go func() {
		for {
			time.Sleep(60 * time.Second)
			rc.DeleteOldData()
		}
	}()

	// Define HTTP handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rc.Increment()
		count := rc.CountRequests()

		// Create a struct to represent the JSON response
		type Response struct {
			Count int `json:"count"`
		}

		// Create an instance of the Response struct with the count value
		jsonResponse := Response{Count: count}

		// Encode the struct into JSON format
		jsonResponseBytes, err := json.Marshal(jsonResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the content type as JSON
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON response back to the client
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonResponseBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Create a new HTTP server with graceful shutdown
	server := &http.Server{
		Addr: ":8080",
	}

	// Start HTTP server in a separate goroutine
	go func() {
		log.Println("Server listening on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server:", err)
		}
	}()

	// Listen for termination signals to gracefully shutdown the server
	// Here the trade-off is in case of safe termination (i.e. ctrl+c or k8s sending SIGTERM)
	// we do not lose the state of the request counter and we will save it's state to the file
	// for the last time. But if a panic happens between the intervals of saving state of request counter
	// to the file, then it's state will be lost.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a timeout to allow in-flight requests to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt to gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Error shutting down server:", err)
	}

	// Attempt to store the last state of request counter
	log.Println("Storing the last state of request counter...")
	err := rc.SaveToFile()
	if err != nil {
		log.Fatal("Could not store the last state of request counter sucessfully", err)
	}

	log.Println("Server gracefully stopped")
}
