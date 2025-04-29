// main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	mode = flag.String("mode", "server", "server or client")
	host = flag.String("host", "127.0.0.1", "server host (client only)")
	port = flag.String("port", "9000", "port to listen on or connect to")
	name = flag.String("name", "anon", "client display name")
)

func main() {
	flag.Parse()
	addr := ":" + *port
	if *mode == "server" {
		startServer(addr)
	} else {
		startClient(*host+":"+*port, *name)
	}
}

// SERVER

var (
	clients   = make(map[net.Conn]bool)
	clientsMu sync.Mutex
)

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
		// skip echoing back to sender if desired
		_, _ = c.Write([]byte(full))
	}
}

func logf(format string, args ...interface{}) {
	ts := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s\n", ts, fmt.Sprintf(format, args...))
}

// CLIENT

func startClient(addr, name string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial error:", err)
		return
	}
	defer conn.Close()
	logf("Connected to %s", addr)

	// receive
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Receive error:", err)
		}
		os.Exit(0)
	}()

	// send
	input := bufio.NewScanner(os.Stdin)
	for {
		if !input.Scan() {
			break
		}
		text := strings.TrimSpace(input.Text())
		if text == "" {
			continue
		}
		msg := fmt.Sprintf("%s: %s", name, text)
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Send error:", err)
			break
		}
	}
}
