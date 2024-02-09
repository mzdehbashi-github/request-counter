# Request Counter

The Request Counter is a simple Go application that tracks the total number of requests received in the past 60 seconds. It provides functionality to increment the request count, count the total number of requests received in the past minute, and save/load the request counter data to/from a file.

## Features

- Increment request count for the current timestamp.
- Count the total number of requests received in the past 60 seconds.
- Save request counter data to a file using gob encoding.
- Load request counter data from a file using gob decoding.
- Automatically delete old data older than 60 seconds.

## Prerequisites

To run this application, ensure you have Go installed on your system.

## Usage

Clone the repository:
```bash
git clone <repository-url>

cd simplesurance-challenge/
go build .
go run .
```

## How it Works
The Request Counter application uses a RequestCounter struct to maintain a counter for the total number of requests received in the past 60 seconds. It utilizes a sync.RWMutex for thread-safe access to the request counter map.

### The main functionalities of the RequestCounter struct include:

- Loading request counter data from a file using gob decoding (when the application starts).
- Incrementing the request count for the current timestamp and return the total count (per HTTP request).
- Counting the total number of requests received in the past 60 seconds(intervally in a separate goroutine).
- Saving request counter data to a file using gob encoding (intervally in a separate goroutine).
- Deleting old data older than 60 seconds (intervally in a separate goroutine).

### File Structure
- main.go: Entry point of the application.
- requestCounter.go: Contains the implementation of the RequestCounter struct and its methods.

