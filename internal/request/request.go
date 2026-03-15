package request

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/RemcoVeens/tcp2http/internal/headers"
)

type Status int

const (
	Initialized Status = iota
	RequestStateParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	Status      Status
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (r *Request, err error) {
	r = &Request{}
	buffer := make([]byte, 8)
	var bytesInBuffer int
	var bytesRead int
	var bytesParsed int

	for {
		if r.Status == Done {
			return r, nil
		}

		if bytesInBuffer == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer[:bytesInBuffer])
			buffer = newBuffer
		}

		n, readErr := reader.Read(buffer[bytesInBuffer:])
		if n > 0 {
			bytesInBuffer += n
			bytesRead += n
		}

		log.Printf("DEBUG main: calling parse with bytesInBuffer=%d", bytesInBuffer)
		parsed, parseErr := r.parse(buffer[:bytesInBuffer])
		if parseErr != nil {
			return nil, parseErr
		}
		if r.Status == Done {
			return r, nil
		}
		if parsed > 0 {
			bytesParsed += parsed
			if parsed >= bytesInBuffer {
				bytesInBuffer = 0
			} else {
				copy(buffer, buffer[parsed:bytesInBuffer])
				bytesInBuffer -= parsed
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				log.Printf("DEBUG main: EOF, r.Status=%d", r.Status)
				if r.Status == Done {
					return r, nil
				}
				if r.Status == ParsingBody && r.Headers.Get("Content-Length") == "" {
					r.Status = Done
					return r, nil
				}
				return nil, fmt.Errorf("connection closed before complete body was received")
			} else {
				return nil, readErr
			}
		}
	}
}

func (r *Request) parse(data []byte) (int, error) {
	if r.Headers == nil {
		r.Headers = headers.NewHeaders()
	}
	switch r.Status {
	case Done:
		return 0, nil
	case Initialized:
		idx := strings.Index(string(data), "\r\n")
		if idx == -1 {
			return 0, nil
		}

		line := string(data[:idx])
		parts := strings.Split(line, " ")
		if len(parts) != 3 {
			return 0, fmt.Errorf("invalid request line: %s", line)
		}
		versionParts := strings.Split(parts[2], "/")
		if len(versionParts) != 2 {
			return 0, fmt.Errorf("invalid http version: %s", parts[2])
		}
		rl := RequestLine{
			Method:        parts[0],
			RequestTarget: parts[1],
			HttpVersion:   versionParts[1],
		}
		if strings.ToUpper(rl.Method) != rl.Method {
			return 0, fmt.Errorf("method is not upper")
		}
		if rl.HttpVersion != "1.1" {
			return 0, fmt.Errorf("http version is not HTTP/1.1: %s", rl.HttpVersion)
		}
		r.RequestLine = rl
		r.Status = RequestStateParsingHeaders
		return idx + 2, nil
	case RequestStateParsingHeaders:
		if len(data) == 0 {
			return 0, fmt.Errorf("no data to parse")
		}
		parsedHeaders, done, parseErr := r.Headers.Parse(data)
		if parseErr != nil {
			return 0, parseErr
		}
		if done {
			r.Status = ParsingBody
			consumed := parsedHeaders
			if parsedHeaders+2 <= len(data) && string(data[parsedHeaders:parsedHeaders+2]) == "\r\n" {
				consumed += 2
			}
			remainingData := len(data) - consumed
			if r.Headers.Get("Content-Length") == "" {
				if remainingData == 0 {
					r.Status = Done
				} else {
					r.Body = append(r.Body, data[consumed:]...)
					r.Status = Done
				}
			}
			return consumed, nil
		}
		return parsedHeaders, nil
	case ParsingBody:
		contentLength := r.Headers.Get("Content-Length")
		log.Printf("DEBUG ParsingBody: len(data)=%d, Content-Length=%s, Status=%d", len(data), contentLength, r.Status)
		if contentLength == "" {
			log.Printf("DEBUG ParsingBody: no Content-Length, appending %d bytes", len(data))
			if len(data) > 0 {
				r.Body = append(r.Body, data...)
			}
			log.Printf("DEBUG ParsingBody: NOT setting Done yet, returning len(data)=%d", len(data))
			return len(data), nil
		}
		ContentLength, _ := strconv.Atoi(contentLength)
		if ContentLength == 0 {
			if len(data) > 0 {
				return 0, fmt.Errorf("body not expected when content length is 0")
			}
			r.Status = Done
			return 0, nil
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > ContentLength {
			return 0, fmt.Errorf("body length (%d) exceeds content length (%d)", len(r.Body), ContentLength)
		} else if len(r.Body) == ContentLength {
			r.Status = Done
			fmt.Println("body parsed:", string(r.Body))
		}
		return len(data), nil

	default:
		return 0, fmt.Errorf("unknown status: %v", r.Status)
	}
}
