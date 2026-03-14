package server

import (
	"bytes"
	"io"

	"github.com/RemcoVeens/tcp2http/internal/request"
	"github.com/RemcoVeens/tcp2http/internal/response"
)

const (
	BadRequestHTML = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

	InternalServerErrorHTML = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

	OKHTML = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
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
		return HandlerError{StatusCode: 400, Message: BadRequestHTML}
	case "/myproblem":
		return HandlerError{StatusCode: 500, Message: InternalServerErrorHTML}
	default:
		w.WriteString(OKHTML)
		return nil
	}
}

func WriteHandlerError(w io.Writer, err HandlerError) error {
	if err := response.WriteStatusLine(w, response.StatusCode(err.StatusCode)); err != nil {
		return err
	}
	headers := response.GetDefaultHTMLHeaders(len(err.Message))
	headers["Connection"] = "close"
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}
	_, wErr := w.Write([]byte(err.Message))
	return wErr
}
