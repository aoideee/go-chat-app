// Server code for a simple TCP chat application
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var (
    port      = flag.String("port", "9000", "port to listen on")
    clients   = make(map[net.Conn]bool)
    clientsMu sync.Mutex
)

func main() {
    flag.Parse()
    addr := ":" + *port
    startServer(addr)
}

func startServer(addr string) {
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Listen error:", err)
        return
    }
    fmt.Println("TCP chat server listening on", addr)
    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Fprintln(os.Stderr, "Accept error:", err)
            continue
        }
        clientsMu.Lock()
        clients[conn] = true
        clientsMu.Unlock()
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer func() {
        clientsMu.Lock()
        delete(clients, conn)
        clientsMu.Unlock()
        conn.Close()
    }()
    remote := conn.RemoteAddr().String()
    logf("CONNECTED: %s", remote)
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        msg := scanner.Text()
        broadcast(remote, msg)
    }
    if err := scanner.Err(); err != nil && err != io.EOF {
        fmt.Fprintln(os.Stderr, "Read error:", err)
    }
    logf("DISCONNECTED: %s", remote)
}

func broadcast(sender, msg string) {
    ts := time.Now().Format(time.RFC3339)
    full := fmt.Sprintf("[%s] %s: %s\n", ts, sender, msg)
    clientsMu.Lock()
    defer clientsMu.Unlock()
    for c := range clients {
        _, _ = c.Write([]byte(full))
    }
}

func logf(format string, args ...interface{}) {
    ts := time.Now().Format(time.RFC3339)
    fmt.Printf("[%s] %s\n", ts, fmt.Sprintf(format, args...))
}