package main

import(
	"fmt"
	"net"
	"bufio"
	"os"
	//"strconv"
)

func main(){
	serverAddress := net.UDPAddr{											//Set the same server address
		Port: 6060,
		IP: net.ParseIP("0.0.0.0"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddress)					//Test server address
	if err != nil{
		panic(err)
	}
	defer conn.Close()
	//localAddress := conn.LocalAddr().(*net.UDPAddr)
	arrivalMessage := "has arrived!"
	_, err = conn.Write([]byte(arrivalMessage))
	if err != nil{
		fmt.Println("Error writing to server: ", err)
	}

	go func(){																//Display received messages
		buf := make([]byte, 1024)
		for{
			n, _, err := conn.ReadFromUDP(buf)
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
		_, err := conn.Write([]byte(text))									//Post message to server
		if err != nil{
			fmt.Println("Error writing to server: ", err)
			break
		}
	}
}
