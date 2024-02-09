package main

import (
	"encoding/gob"
	"log"
	"os"
	"sync"
	"time"
)

// RequestCounter maintains a counter for the total number of requests received in the past 60 seconds.
type RequestCounter struct {
	Requests map[int64]int // Map to store request counts with timestamp
	Mutex    sync.RWMutex  // ReadWrite Mutex to ensure read and write safety
	filename string
}

// NewRequestCounter initializes a new RequestCounter.
func NewRequestCounter(filename string) *RequestCounter {
	rc := &RequestCounter{
		Requests: make(map[int64]int),
		filename: filename,
	}

	// Load previous data from file if available
	if _, err := os.Stat(filename); err == nil {
		if err := rc.LoadFromFile(); err != nil {
			log.Fatal("Error loading data from file:", err)
		}
	}

	return rc
}

// Increment increments the request count for the current timestamp.
func (rc *RequestCounter) Increment() {
	currentTimestamp := time.Now().Unix()
	rc.Mutex.Lock()
	defer rc.Mutex.Unlock()
	rc.Requests[currentTimestamp]++
}

// CountRequests counts the total number of requests received in the past 60 seconds.
func (rc *RequestCounter) CountRequests() int {
	currentTimestamp := time.Now().Unix()
	count := 0
	rc.Mutex.RLock()
	defer rc.Mutex.RUnlock()
	for timestamp, reqCount := range rc.Requests {
		if timestamp >= currentTimestamp-60 {
			count += reqCount
		}
	}
	return count
}

// SaveToFile saves the request counter data to a file using gob encoding.
func (rc *RequestCounter) SaveToFile() error {
	file, err := os.Create(rc.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(rc.Requests); err != nil {
		return err
	}
	return nil
}

// LoadFromFile loads the request counter data from a file using gob decoding.
func (rc *RequestCounter) LoadFromFile() error {
	file, err := os.Open(rc.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&rc.Requests); err != nil {
		return err
	}
	return nil
}

func (rc *RequestCounter) DeleteOldData() {
	rc.Mutex.Lock()
	expiredTimestamp := time.Now().Unix() - 61
	for timestamp := range rc.Requests {
		if timestamp < expiredTimestamp {
			delete(rc.Requests, timestamp)
		}
	}
	rc.Mutex.Unlock()
}
