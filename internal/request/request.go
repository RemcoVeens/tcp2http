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

		parsed, parseErr := r.parse(buffer[:bytesInBuffer])
		if parseErr != nil {
			return nil, parseErr
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
				if r.Status == Done {
					_ = bytesRead
					_ = bytesParsed
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
		fmt.Println(r)
		return idx + 2, nil
	case RequestStateParsingHeaders:
		if len(data) == 0 {
			return 0, fmt.Errorf("no data to parse")
		}
		fmt.Println(string(data))
		parsedHeaders, done, parseErr := r.Headers.Parse(data)
		if parseErr != nil {
			log.Println("error")
			return 0, parseErr
		}
		if done {
			log.Println("headers parsed:", r.Headers)
			r.Status = ParsingBody
			if parsedHeaders < len(data) {
				data = data[parsedHeaders:]
				if len(data) >= 2 && string(data[:2]) == "\r\n" {
					parsedHeaders += 2
				}
			}
		}
		return parsedHeaders, nil
	case ParsingBody:
		log.Println("parsing body")
		contentLength := r.Headers.Get("Content-Length")
		if contentLength == "" {
			if len(data) == 0 {
				r.Status = Done
				return len(data), nil
			}
			r.Body = append(r.Body, data...)
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
