//client.go
 
package main

import(
	"fmt"
	"flag"
	"net"
	"bufio"
	"os"
	"time"
)

var startTime time.Time

func main(){

	name := flag.String("name", "Unknown_Client", "Client name")
	flag.Parse()

	serverAddress := net.UDPAddr{											//Set the same server address
		Port: 6060,
		IP: net.ParseIP("0.0.0.0"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddress)					//Test server address
	if err != nil{
		panic(err)
	}
	defer conn.Close()
	fmt.Println("You have joined the chat and may start messaging.")
	//localAddress := conn.LocalAddr().(*net.UDPAddr)
	arrivalMessage := *name + " has arrived!"
	_, err = conn.Write([]byte(arrivalMessage))
	if err != nil{
		fmt.Println("Error writing to server: ", err)
	}

	go func(){																//Display received messages
		buf := make([]byte, 1024)
		for{
			n, _, err := conn.ReadFromUDP(buf)
			timeTaken := time.Since(startTime).Seconds()
			precision := fmt.Sprintf("%.6f", timeTaken)
			fmt.Println("Round Trip time: ", precision)
			if err != nil{
				fmt.Println("Error reading from server: ", err)
				return
			}
			fmt.Println(string(buf[:n]))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan(){
		text := scanner.Text()
		if text == ""{														//Don't write post message if empty
			continue
		}
		text = *name + ": " + text
		_, err := conn.Write([]byte(text))									//Post message to server
		startTime = time.Now()
		if err != nil{
			fmt.Println("Error writing to server: ", err)
			break
		}
	}
}

func timestamp() string {
	t := time.FixedZone("America/Chicago (No DST)", -6*60*60)
	return time.Now().In(t).Format("[15:04:05.000000]")
}