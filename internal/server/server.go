package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/RemcoVeens/tcp2http/internal/response"
)

type Server struct {
	lisener net.Listener
	open    atomic.Bool
}

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := Server{l, atomic.Bool{}}
	go server.listen()
	return &server, nil
}

func (s Server) Close() error {
	if s.lisener == nil {
		return fmt.Errorf("server not started")
	}
	return s.lisener.Close()
}

func (s *Server) listen() {
	s.open.Store(true)
	for {
		conn, err := s.lisener.Accept()
		if err != nil {
			if !s.open.Load() {
				return
			}
			log.Printf("accept error: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s Server) handle(conn net.Conn) {
	defer conn.Close()
	response.WriteStatusLine(conn, response.OK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	// log.Printf("got a connection from %s", conn.RemoteAddr())

}
