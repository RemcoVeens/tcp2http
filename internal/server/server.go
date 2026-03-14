package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/RemcoVeens/tcp2http/internal/request"
	"github.com/RemcoVeens/tcp2http/internal/response"
)

type Server struct {
	lisener net.Listener
	handler Handler
	open    atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := Server{l, handler, atomic.Bool{}}
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		WriteHandlerError(conn, HandlerError{StatusCode: response.BadRequest, Message: err.Error()})
		return
	}

	buf := &bytes.Buffer{}
	err = s.handler(buf, req)
	if err != nil {
		if handlerErr, ok := err.(HandlerError); ok {
			WriteHandlerError(conn, handlerErr)
			return
		}
		WriteHandlerError(conn, HandlerError{StatusCode: response.InteranlServerError, Message: err.Error()})
		return
	}

	response.WriteStatusLine(conn, response.OK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(buf.Len()))
	conn.Write(buf.Bytes())
}
