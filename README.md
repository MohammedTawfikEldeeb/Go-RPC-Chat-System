# Go-RPC-Chat-System

A simple multi-client chat app built in Go using RPC, goroutines, channels, and mutex.

## Files
- `server.go` – Chat server
- `client.go` – Chat client

## Features
- RPC-based communication  
- Shared chat history for new clients  
- Unique user IDs  
- Join/leave notifications  
- Message broadcast (no self-echo)  
- Concurrent clients with goroutines  
- Thread-safe with mutex  

## Run
```bash
go run server.go
go run client.go
