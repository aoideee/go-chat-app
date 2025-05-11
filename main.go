// main.go
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mode         = flag.String("mode", "server", "server or client")
	host         = flag.String("host", "127.0.0.1", "server host (client only)")
	port         = flag.String("port", "9000", "port to listen on or connect to")
	name         = flag.String("name", "anon", "client display name")

	// test mode flags
	testMode     = flag.Bool("test", false, "run automated test sequence (client only)")
	testCount    = flag.Int("count", 100, "number of messages to send in test mode")
	testInterval = flag.Duration("interval", 1*time.Second, "interval between test messages")
	timeout      = flag.Duration("timeout", 5*time.Second, "read timeout per message")
	csvPath      = flag.String("csv", "metrics.csv", "output CSV file for test metrics")
)

func main() {
	flag.Parse()
	addr := ":" + *port

	if *mode == "server" {
		startServer(addr)
		return
	}

	// client path
	if *testMode {
		runTest(*host+":"+*port, *name)
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
	for input.Scan() {
		text := strings.TrimSpace(input.Text())
		if text == "" {
			continue
		}
		msg := fmt.Sprintf("%s: %s", name, text)
		if _, err := conn.Write([]byte(msg + "\n")); err != nil {
			fmt.Fprintln(os.Stderr, "Send error:", err)
			break
		}
	}
}

// TEST MODE

type metric struct {
	seq      int
	rttMs    int64
	lost     bool
	orderErr bool
}

func runTest(addr, name string) {
	fmt.Printf("Running test mode: %d messages to %s every %s (timeout %s)\n",
		*testCount, addr, testInterval.String(), timeout.String())

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial error:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var metrics []metric
	expected := 1

	for i := 1; i <= *testCount; i++ {
		// prepare and send
		t0 := time.Now()
		payload := fmt.Sprintf("%d|%d", i, t0.UnixNano())
		if _, err := conn.Write([]byte(payload + "\n")); err != nil {
			fmt.Fprintf(os.Stderr, "Send error at #%d: %v\n", i, err)
			metrics = append(metrics, metric{i, 0, true, false})
			continue
		}

		// wait for response or timeout
		conn.SetReadDeadline(time.Now().Add(*timeout))
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Timeout/loss at #%d\n", i)
			metrics = append(metrics, metric{i, 0, true, false})
		} else {
			// parse out "seq|sendNano" from the broadcast message
			parts := strings.SplitN(line, ": ", 3)
			if len(parts) < 3 {
				metrics = append(metrics, metric{i, 0, false, true})
			} else {
				recv := strings.TrimSpace(parts[2])
				sub := strings.Split(recv, "|")
				if len(sub) != 2 {
					metrics = append(metrics, metric{i, 0, false, true})
				} else {
					seqRecv, err1 := strconv.Atoi(sub[0])
					sendNano, err2 := strconv.ParseInt(sub[1], 10, 64)
					if err1 != nil || err2 != nil {
						metrics = append(metrics, metric{i, 0, false, true})
					} else {
						rtt := time.Since(time.Unix(0, sendNano))
						orderErr := seqRecv != expected
						metrics = append(metrics, metric{i, rtt.Milliseconds(), false, orderErr})
						expected = seqRecv + 1
					}
				}
			}
		}

		time.Sleep(*testInterval)
	}

	// write CSV
	file, err := os.Create(*csvPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CSV create error:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"sequence", "rtt_ms", "lost", "order_error"})
	for _, m := range metrics {
		lost := "0"
		if m.lost {
			lost = "1"
		}
		ord := "0"
		if m.orderErr {
			ord = "1"
		}
		writer.Write([]string{
			strconv.Itoa(m.seq),
			strconv.FormatInt(m.rttMs, 10),
			lost,
			ord,
		})
	}

	fmt.Println("Test complete — metrics written to", *csvPath)
}