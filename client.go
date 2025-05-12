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

	name := flag.String("name", "Unknown_Client", "Client name")			//Name Flag
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
	arrivalMessage := *name + " has arrived!"
	_, err = conn.Write([]byte(arrivalMessage))
	if err != nil{
		fmt.Println("Error writing to server: ", err)
	}

	go func(){																//Display received messages
		buf := make([]byte, 1024)
		for{
			n, _, err := conn.ReadFromUDP(buf)
			// timeTaken := time.Since(startTime).Seconds()		//Round Trip Testing
			// precision := fmt.Sprintf("%.6f", timeTaken)		//Round Trip Testing
			// fmt.Println("Round Trip time: ", precision)		//Round Trip Testing
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
		if text == ""{														//Don't write message if empty
			continue
		}
		text = *name + ": " + text
		_, err := conn.Write([]byte(text))									//Post message to server
		// startTime = time.Now()		//Round Trip Testing
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


// Use this for packet order testing
// for i := 0; i <= 9; i++{
// 	newText := text + fmt.Sprintf("%d", i)
// 	_, err := conn.Write([]byte(newText))
// 	time.Sleep(10 * time.Millisecond)
// 	// startTime = time.Now()		//Round Trip Testing
// 	if err != nil{
// 		fmt.Println("Error writing to server: ", err)
// 		break
// 	}
// }