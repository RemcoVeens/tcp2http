package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func getLinesChannel(f net.Conn) <-chan string {
	ch := make(chan string)
	// defer close(ch)
	go func() {
		var line string
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading file:", err)
				break
			}
			parts := strings.Split(string(buffer[:n]), "\n")
			line += parts[0]
			if len(parts) > 1 {
				ch <- fmt.Sprintf("%s\n", line)
				line = parts[1]
			}
		}
		close(ch)
		fmt.Println("listener closed")
	}()
	return ch
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
		return
	}
	defer conn.Close()
	fmt.Println("listener accepted")
	ch := getLinesChannel(conn)
	for line := range ch {
		fmt.Printf("read: %s", line)
	}
}
