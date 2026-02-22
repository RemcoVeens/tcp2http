package main

import (
	"fmt"
	"net"

	"github.com/RemcoVeens/tcp2http/internal/request"
)

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
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	fmt.Printf(
		"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
}
