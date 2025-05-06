//server.go

package main

import(
	"fmt"
	"net"
	"time"
	"strings"
)

func main(){
	address := net.UDPAddr{																//Set the server port
		Port: 6060,
		IP: net.ParseIP("0.0.0.0"),														// //Port and IP are both needed in UDP
	}

	listener, err := net.ListenUDP("udp", &address)										//Create the listener
	if err != nil{
		panic(err)
	}
	defer listener.Close()
	fmt.Println(timestamp(), "UDP chat server listening on localhost: ", address.String())

	clients := make(map[string]*net.UDPAddr)											//Array that will store the address of every message sent
	buf := make([]byte, 1024)

	for{																				//Loop to accept messages
		n, clientAddress, err := listener.ReadFromUDP(buf)
		if err != nil{
			fmt.Println("Error reading from client: ", err)
			continue																	// //Continue starts a new iteration of the loop
		}

		mail := string(buf[:n])
		split := strings.SplitN(mail, " ", 2)

		name := split[0]
		message := split[1]

		if message == "has arrived!"{
			fmt.Printf("%s %s %s\n", timestamp(), name, message)		//clientAddress.String()
			message = name + " " + message
		}else{
			fmt.Printf("%s %s: %s\n", timestamp(), name, message)							//Print client messages
			message = name + ": " + message
		}

		clients[clientAddress.String()] = clientAddress

		for _, address := range clients {												//Write messages to all clients
			if address.String() != clientAddress.String(){
				if _, err := listener.WriteToUDP([]byte(message), address); err != nil{
					fmt.Println("Error writing to client ", address.String(), ":", err)
				}
			}
		}
	}
}

func timestamp() string {
	t := time.FixedZone("America/Chicago (No DST)", -6*60*60)
	return time.Now().In(t).Format("[15:04:05]")
}

//Advancements: could add client timeout and removal

//Test if order is always correct
//Test speed
//Test server sending to wrong clients or vice versa