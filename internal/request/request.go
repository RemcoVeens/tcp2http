package request

import (
	"fmt"
	"io"
	"strings"
)

type Status int

const (
	Initialized Status = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	Status      Status
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
			copy(buffer, buffer[parsed:bytesInBuffer])
			bytesInBuffer -= parsed
		}

		if readErr != nil {
			if readErr == io.EOF {
				if r.Status == Done || bytesInBuffer == 0 {
					_ = bytesRead
					_ = bytesParsed
					return r, nil
				}
			} else {
				return nil, readErr
			}
		}
	}
}

func (r *Request) parse(data []byte) (int, error) {
	if r.Status == Done {
		return 0, nil
	}

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
	r.Status = Done

	return idx + 2, nil
}
