package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestHttpHandler_CountRequests(t *testing.T) {
	// Scenario: First send 10 concurrent requests to the endpoint, some of them will
	// get timeout error, since the server can only process limited amount of concurrent requests.
	// then send another request to get the final count of requests (successfully processed request)
	// the expected value should be the number of successfully processed requests + 1 (the last request)

	// Do preparation
	filename := "test_request_counter.gob"
	rc := NewRequestCounter(filename, rcWindowDuration)
	httpHandler := &HttpHandler{
		rc:             rc,
		concurrentReqs: make(chan struct{}, maxConcurrentRequests),
	}

	// Simulate sending 10 concurrent requests
	var wg sync.WaitGroup

	wg.Add(10)
	statusCodesChan := make(chan int, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Errorf("Error creating request: %v", err)
				return
			}

			// Create a new HTTP request recorder,
			// for every call to the handler.
			recorder := httptest.NewRecorder()
			httpHandler.CountRequests(recorder, req)
			statusCodesChan <- recorder.Code
		}()
	}

	wg.Wait()

	// count all succesfull requests
	successfullRequestCount := 0
	for i := 0; i < 10; i++ {
		statusCode := <-statusCodesChan
		if statusCode == http.StatusOK {
			successfullRequestCount++
		} else {
			// raise error, since we do not expect any other failure cases other than timeout
			if statusCode != http.StatusRequestTimeout {
				t.Errorf("expected timeout status, but got %d", statusCode)
			}
		}
	}

	// Send the last request to check the final count value
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	reqRecorder := httptest.NewRecorder()
	httpHandler.CountRequests(reqRecorder, req)

	// Check the response status code
	if reqRecorder.Code != http.StatusOK {
		t.Errorf("Expected status OK; got %d", reqRecorder.Code)
	}

	// Decode the response JSON
	var response struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(reqRecorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Error decoding JSON response: %v", err)
	}

	// Check if the count value is as expected
	expectedCount := successfullRequestCount + 1
	if response.Count != expectedCount {
		t.Errorf("Expected count %d; got %d", expectedCount, response.Count)
	}
}
