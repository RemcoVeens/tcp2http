package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.ResolveUDPAddr("udp", "127.0.0.1:42069")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, listener)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer conn.Close()
	buffer := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		str, err := buffer.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		conn.Write([]byte(str))
	}
}
