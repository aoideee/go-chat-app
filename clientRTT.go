// client.go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var (
	startTimes = make(map[string]time.Time)
	mu         sync.Mutex
)

func main() {
	name := flag.String("name", "Unknown_Client", "Client name")
	flag.Parse()

	serverAddress := net.UDPAddr{
		Port: 6060,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("You have joined the chat and may start messaging.")

	// Notify server of arrival
	arrivalMessage := *name + " has arrived!"
	_, err = conn.Write([]byte(arrivalMessage))
	if err != nil {
		fmt.Println("Error writing to server: ", err)
	}

	// Channel to coordinate reply receipt
	replyReceived := make(chan string)

	// Goroutine to handle incoming messages
	go func() {
		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error reading from server:", err)
				return
			}
			received := string(buf[:n])
			id := received[len(received)-3:]

			mu.Lock()
			start, ok := startTimes[id]
			mu.Unlock()

			if ok {
				rtt := time.Since(start).Seconds()
				fmt.Printf("Round Trip for msg %s: %.6f seconds\n", id, rtt)
				replyReceived <- id // notify sender
			} else {
				fmt.Println("Received:", received)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		text = *name + ": " + text

		for i := 0; i < 100; i++ {
			msgID := fmt.Sprintf("%03d", i)
			newText := text + msgID

			mu.Lock()
			startTimes[msgID] = time.Now()
			mu.Unlock()

			_, err := conn.Write([]byte(newText))
			if err != nil {
				fmt.Println("Error writing to server:", err)
				break
			}

			// Wait for reply before continuing
			<-replyReceived
		}
	}
}
