package main

import (
	"sync"
	"testing"
	"time"
)

func TestRequestCounter_Increment(t *testing.T) {
	rc := &RequestCounter{
		Requests:       make(map[int64]int),
		windowDuration: 2 * time.Second,
	}

	// Increment the request count
	rc.Increment()

	// Verify that the request count has been incremented
	count := rc.CountRequests()
	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}
}

func TestRequestCounter_CountRequests(t *testing.T) {
	rc := &RequestCounter{
		Requests:       make(map[int64]int),
		windowDuration: 1 * time.Second,
	}

	// Increment the request count
	rc.Increment()

	// Sleep for a while to simulate elapsed time
	time.Sleep(2 * time.Second)

	// Verify that the count includes only requests within the window duration
	count := rc.CountRequests()
	if count != 0 {
		t.Errorf("Expected count to be 0, got %d", count)
	}
}

func TestRequestCounter_DeleteOldData(t *testing.T) {
	rc := &RequestCounter{
		Requests:       make(map[int64]int),
		windowDuration: 1 * time.Second,
	}

	// Increment the request count
	rc.Increment()

	// Sleep for a while to simulate elapsed time
	time.Sleep(2 * time.Second)

	// Delete old data
	rc.DeleteOldData()

	if requestsCount := len(rc.Requests); requestsCount != 0 {
		t.Errorf("Expected rc.Requests to be empty, got %d entries", requestsCount)
	}
}

func TestConcurrency(t *testing.T) {
	rc := &RequestCounter{
		Requests:       make(map[int64]int),
		windowDuration: rcWindowDuration,
	}

	// Create multiple goroutines to increment the counter concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rc.Increment()
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify that the count matches the expected value
	count := rc.CountRequests()
	if count != 100 {
		t.Errorf("Expected count to be 100, got %d", count)
	}
}
