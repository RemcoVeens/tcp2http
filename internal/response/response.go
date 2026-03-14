package response

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/RemcoVeens/tcp2http/internal/headers"
)

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

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for hdr := range headers {
		log.Printf("%s: %s\r\n", hdr, headers[hdr])
		_, err := w.Write([]byte(hdr + ": " + headers[hdr] + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}
