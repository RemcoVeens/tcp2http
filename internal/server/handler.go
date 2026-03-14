package server

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/RemcoVeens/tcp2http/internal/headers"
	"github.com/RemcoVeens/tcp2http/internal/request"
	"github.com/RemcoVeens/tcp2http/internal/response"
)

type HandlerError struct {
	StatusCode int
	Message    string
}
type Handler func(w *bytes.Buffer, r *request.Request) error

func (e HandlerError) Error() string {
	return e.Message
}

func Handle(w *bytes.Buffer, r *request.Request) error {
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		return HandlerError{StatusCode: 400, Message: "Your problem is not my problem\n"}
	case "/myproblem":
		return HandlerError{StatusCode: 500, Message: "Woopsie, my bad\n"}
	default:
		w.WriteString("All good, frfr\n")
		return nil
	}
}

type StatusCode int

const (
	OK                  = 200
	BadRequest          = 400
	InteranlServerError = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case OK:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return err
	case BadRequest:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return err
	case InteranlServerError:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return err
	default:
		_, err := w.Write([]byte(fmt.Sprintf("HTTP/1.1 %d\r\n", statusCode)))
		return err
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.Headers{}
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"
	return headers
}

func WriteHandlerError(w io.Writer, err HandlerError) error {
	if err := response.WriteStatusLine(w, response.StatusCode(err.StatusCode)); err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(err.Message))
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}
	_, wErr := w.Write([]byte(err.Message))
	return wErr
}
