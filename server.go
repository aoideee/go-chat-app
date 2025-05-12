//server.go

package main

import(
	"fmt"
	"net"
	"time"
	"strings"
	"sync"
)

type client struct{
	address *net.UDPAddr
	lastSeen time.Time
}

var mutex sync.Mutex 

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

	clients := make(map[string]client)											//Array that will store the address of every message sent
	buf := make([]byte, 1024)



	go func (){
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
	
		for range ticker.C{
			for k, v := range clients{
				if time.Now().Sub(v.lastSeen) > (3 * time.Minute){
					// fmt.Println(v.address.String(), " has been inactive for 3 minutes and will be removed.")
					message := "You have been inactive for 3 minutes and will be removed."
					if _, err := listener.WriteToUDP([]byte(message), v.address); err != nil{
						fmt.Println("Error writing to client ", address.String(), ":", err)
					}
					mutex.Lock()
					delete(clients, k)
					mutex.Unlock()
				}
			}
		}
	}()



	for{																				//Loop to accept messages
		n, clientAddress, err := listener.ReadFromUDP(buf)
		if err != nil{
			fmt.Println("Error reading from client: ", err)
			continue																	// //Continue starts a new iteration of the loop
		}

		message := string(buf[:n])
		split := strings.SplitN(message, " ", 2)

		mail := split[1]
		
		if mail == "has arrived!"{
			fmt.Printf("%s  -  %s\n", timestamp(), message)	
			message = timestamp() + "  -  " + message
		}else{
			fmt.Printf("%s  -  %s\n", timestamp(), message)							//Print client messages
			message = timestamp() + "  -  " + message
		}

		mutex.Lock()
		clients[clientAddress.String()] = client{
			address: clientAddress,
			lastSeen: time.Now(),
		}
		mutex.Unlock()	

		for _, info := range clients {													//Write messages to all clients
			if info.address != clientAddress{
				if _, err := listener.WriteToUDP([]byte(message), info.address); err != nil{
					fmt.Println("Error writing to client ", info.address.String(), ":", err)
				}
			}
		}
	}
}

func timestamp() string {
	t := time.FixedZone("America/Chicago (No DST)", -6*60*60)
	return time.Now().In(t).Format("[15:04:05.000000]")
}



// Use this to check which clients the server has stored
// for _, info := range clients {
// 	fmt.Println(info.address.String(), info.lastSeen)
// }	