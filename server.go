//main.go

package main

import(
	"fmt"
	"net"
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
	fmt.Println("Server listening on localhost: ", address.String())

	clients := make(map[string]*net.UDPAddr)											//Array that will store the address of every message sent
	buf := make([]byte, 1024)

	for{																				//Loop to accept messages
		n, clientAddress, err := listener.ReadFromUDP(buf)
		if err != nil{
			fmt.Println("Error reading from client: ", err)
			continue																	// //Continue starts a new iteration of the loop
		}

		message := string(buf[:n])
		if message == "has arrived!"{
			fmt.Printf("%s %s\n", clientAddress.String(), message)	
			message = clientAddress.String() + " " + message
		}else{
			fmt.Printf("%s: %s\n", clientAddress.String(), message)							//Print client messages
			message = clientAddress.String() + ": " + message
		}

		clients[clientAddress.String()] = clientAddress

		for _, address := range clients {												//Write messages to all clients
			if address != clientAddress{
				if _, err := listener.WriteToUDP([]byte(message), address); err != nil{
					fmt.Println("Error writing to client ", address.String(), ":", err)
				}
			}
		}
	}
}

//Advancements: could add client timeout and removal

//Test if order is always correct
//Test speed
//Test server sending to wrong clients or vice versa