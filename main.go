package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const saveToFileDuration = 10 * time.Second
const removeOldKeysDuration = 60 * time.Second

func savePeriodically(ctx context.Context, rc *RequestCounter, wg *sync.WaitGroup, duration time.Duration) {
	defer wg.Done()
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rc.SaveToFile()
		case <-ctx.Done():
			log.Println("Saving routine cancelled, saving data for the last time")
			// Save one last time before exiting
			rc.SaveToFile()
			return
		}
	}
}

func removeOldKeysPeriodically(ctx context.Context, rc *RequestCounter, wg *sync.WaitGroup, duration time.Duration) {
	defer wg.Done()
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rc.DeleteOldData()
		case <-ctx.Done():
			log.Println("Removing old keys routine cancelled")
			return
		}
	}
}

func startHTTPServer(ctx context.Context, server *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()

	go func() {
		log.Println("Server listening on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server:", err)
		}
	}()

	// Listen for termination signals to gracefully shutdown the server
	<-ctx.Done()

	// Attempt to gracefully shutdown the server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Error shutting down server:", err)
	}

	log.Println("Server gracefully stopped")
}

func main() {
	filename := "request_counter.gob"
	rc := NewRequestCounter(filename)

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

	// Create a new HTTP server
	server := &http.Server{
		Addr: ":8080",
	}

	// Context for managing graceful shutdown of all goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// Wait group to make sure all goroutines are getting terminated,
	// before the main thread is terminated.
	var wg sync.WaitGroup

	// Increment the WaitGroup for each goroutine:
	// - savePeriodically
	// - removeOldKeysPeriodically
	// - startHTTPServer
	wg.Add(3)
	go savePeriodically(ctx, rc, &wg, saveToFileDuration)
	go removeOldKeysPeriodically(ctx, rc, &wg, removeOldKeysDuration)
	go startHTTPServer(ctx, server, &wg)

	// Listen for termination signals to cancel the context
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signals or cancellation of context
	<-signalCh
	log.Println("Received termination signal, cancelling context...")
	cancel()

	// Wait for all goroutines to finish gracefully
	wg.Wait()
	log.Println("main gracefully stopped")
}
