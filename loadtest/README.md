# Load Test Tool

This is a simple load testing tool written in Go. It sends a configurable number of HTTP requests to the server and collects and counts the status codes of the responses.


## How to run:
Please make sure that the server is running lcoally on prot 8080 and the run this command in a new terminal:
```
go run main.go -n <number of concurrent requests: int, default=100>
```