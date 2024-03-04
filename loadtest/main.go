package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Parse command line flags
	var numRequests int
	flag.IntVar(&numRequests, "n", 100, "Number of requests to send")
	flag.Parse()

	// Create a channel to collect status codes
	statusCodes := make(chan int, numRequests)

	// Prepare requests
	request, err := http.NewRequest("GET", "http://localhost:8080", nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}

	// Send requests
	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := http.Client{
				Timeout: time.Second * 5, // Timeout for each request
			}
			resp, err := client.Do(request)
			if err != nil {
				log.Println("Request error:", err)
				statusCodes <- 0 // Increment count for unknown errors
				return
			}
			defer resp.Body.Close()

			// Send status code into the channel
			statusCodes <- resp.StatusCode
		}()
	}

	// Wait for all requests to finish
	wg.Wait()
	close(statusCodes)

	// Count status codes
	statusCodeCounts := make(map[int]int)
	for code := range statusCodes {
		statusCodeCounts[code]++
	}

	// Print results
	log.Println("Load test results:")
	log.Printf("Total requests: %d\n", numRequests)
	log.Printf("Successful requests: %d\n", statusCodeCounts[http.StatusOK])
	log.Printf("Timeout requests: %d\n", statusCodeCounts[http.StatusRequestTimeout])
	log.Println("Other status codes:")
	for code, count := range statusCodeCounts {
		if code != http.StatusOK && code != http.StatusRequestTimeout && count > 0 {
			log.Printf("%d: %d\n", code, count)
		}
	}
}
