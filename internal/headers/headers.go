package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if strings.HasPrefix(string(data), "\r\n") {
		return n + 2, true, nil
	}
	idx := strings.Index(string(data), "\r\n")
	if idx == -1 {
		return n, false, nil
	}
	line := string(data[:idx])
	parts := strings.Split(line, ":")
	if len(parts) >= 2 {
		head := strings.ToLower(parts[0])
		if strings.Contains(head, " ") {
			return n, false, fmt.Errorf("invalid header: %s", head)
		}
		for _, c := range head {
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", c) {
				continue
			}
			return n, false, fmt.Errorf("invalid header: %s", head)
		}
		value := strings.TrimSpace(strings.Join(parts[1:], ":"))
		if _, ok := h[head]; ok {
			h[head] += ", " + value
		} else {
			h[head] = value
		}
	}
	n += idx + 2
	data = data[idx+2:]
	return
}

func NewHeaders() Headers {
	return Headers{}
}
