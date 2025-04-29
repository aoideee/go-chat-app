# Go TCP Chat App

A minimal multi-client TCP chat using Go.  
Supports reliable, ordered messaging with timestamped logs.

## Prerequisites

- Go 1.18+ installed

## Running

### Server

```bash
go run main.go -mode server -port 9000
```

### Client

```bash
go run main.go -mode client -host 127.0.0.1 -port 9000 -name Tysha
```

and

```bash
go run main.go -mode client -host 127.0.0.1 -port 9000 -name Ashley
```

## Features
- Broadcast chat: any message from one client is delivered to all connected clients

- Timestamps on every send, receive, connect, and disconnect event

- Simple CLI: type your message and press Enter

## Demo Steps
1. Start the server.

2. Open two (or more) terminals and launch clients with different -name flags.

3. Send messages; observe timestamps and ordering.

4. Simulate packet loss/reordering in tests as needed.
