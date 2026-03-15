package server

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RemcoVeens/tcp2http/internal/headers"
	"github.com/RemcoVeens/tcp2http/internal/request"
	"github.com/RemcoVeens/tcp2http/internal/response"
)

func writeChunk(w io.Writer, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(w, "%x\r\n", len(data)); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}

func writeChunkedEnd(w io.Writer) error {
	_, err := w.Write([]byte("0\r\n\r\n"))
	return err
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

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
type Handler func(w io.Writer, r *request.Request) error

func (e HandlerError) Error() string {
	return e.Message
}

func Handle(w io.Writer, r *request.Request) error {
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		return HandlerError{StatusCode: 400, Message: BadRequestHTML}
	case "/myproblem":
		return HandlerError{StatusCode: 500, Message: InternalServerErrorHTML}
	case "/httpbin/html":
		url := "https://httpbin.org/html"
		resp, err := httpClient.Get(url)
		if err != nil {
			return HandlerError{StatusCode: 502, Message: fmt.Sprintf("error fetching upstream: %v", err)}
		}
		defer resp.Body.Close()

		response.WriteStatusLine(w, response.OK)
		hdrs := headers.Headers{}
		hdrs["Transfer-Encoding"] = "chunked"
		hdrs["Trailer"] = "X-Content-Sha256, X-Content-Length"
		hdrs["Content-Type"] = "text/html"
		response.WriteHeaders(w, hdrs)

		var fullBody []byte
		buffer := make([]byte, 1024)
		for {
			n, readErr := resp.Body.Read(buffer)
			if n > 0 {
				fullBody = append(fullBody, buffer[:n]...)
				if wErr := writeChunk(w, buffer[:n]); wErr != nil {
					return HandlerError{StatusCode: 502, Message: fmt.Sprintf("error writing response: %v", wErr)}
				}
				if flusher, ok := w.(interface{ Flush() }); ok {
					flusher.Flush()
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					writeChunkedEnd(w)
					hash := sha256.Sum256(fullBody)
					w.Write([]byte("X-Content-Sha256: " + fmt.Sprintf("%x", hash) + "\r\n"))
					w.Write([]byte("X-Content-Length: " + fmt.Sprintf("%d", len(fullBody)) + "\r\n"))
					if flusher, ok := w.(interface{ Flush() }); ok {
						flusher.Flush()
					}
					time.Sleep(100 * time.Millisecond)
					return nil
				}
				return HandlerError{StatusCode: 502, Message: fmt.Sprintf("error reading response: %v", readErr)}
			}
		}
	default:
		target := r.RequestLine.RequestTarget
		if subpath, ok := strings.CutPrefix(target, "/httpbin/"); ok {
			url := fmt.Sprintf("https://httpbin.org/%s", subpath)
			resp, err := httpClient.Get(url)
			if err != nil {
				return HandlerError{StatusCode: 502, Message: fmt.Sprintf("error fetching upstream: %v", err)}
			}
			defer resp.Body.Close()

			response.WriteStatusLine(w, response.OK)
			hdrs := headers.Headers{}
			hdrs["Transfer-Encoding"] = "chunked"
			hdrs[`Trailer`] = `X-Content-Sha256, X-Content-Length`
			hdrs["Content-Type"] = "text/html"
			response.WriteHeaders(w, hdrs)

			var fullBody []byte
			buffer := make([]byte, 1024)
			for {
				n, readErr := resp.Body.Read(buffer)
				if n > 0 {
					fullBody = append(fullBody, buffer[:n]...)
					if wErr := writeChunk(w, buffer[:n]); wErr != nil {
						return HandlerError{StatusCode: 502, Message: fmt.Sprintf("error writing response: %v", wErr)}
					}
					if flusher, ok := w.(interface{ Flush() }); ok {
						flusher.Flush()
					}
				}
				if readErr != nil {
					if readErr == io.EOF {
						_, err := w.Write([]byte("0\r\n"))
						if err != nil {
							return HandlerError{StatusCode: 502, Message: fmt.Sprintf("error writing chunk end: %v", err)}
						}
						hash := sha256.Sum256(fullBody)
						w.Write([]byte("X-Content-SHA256: " + fmt.Sprintf("%x", hash) + "\r\n"))
						w.Write([]byte("X-Content-Length: " + fmt.Sprintf("%d", len(fullBody)) + "\r\n"))
						w.Write([]byte("\r\n"))
						if flusher, ok := w.(interface{ Flush() }); ok {
							flusher.Flush()
						}
						time.Sleep(100 * time.Millisecond)
						return nil
					}
					return HandlerError{StatusCode: 502, Message: fmt.Sprintf("error reading response: %v", readErr)}
				}
			}
		}

		w.Write([]byte(OKHTML))
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
