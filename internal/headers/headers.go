package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (int, bool, error) {
	var n int
	for len(data) > 0 {
		if len(data) >= 2 && string(data[:2]) == "\r\n" {
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
		if len(data) == 0 {
			return n, false, nil
		}
		if len(data) >= 2 && string(data[:2]) == "\r\n" {
			return n, true, nil
		}
	}
	return n, false, nil
}
func (h Headers) Get(key string) string {
	key = strings.ToLower(key)
	if value, ok := h[key]; ok {
		return value
	}
	return ""
}
func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func NewHeaders() Headers {
	return Headers{}
}
func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}
