package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	// defer close(ch)
	go func() {
		var line string
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading file:", err)
				break
			}
			parts := strings.Split(string(buffer[:n]), "\n")
			line += parts[0]
			if len(parts) > 1 {
				ch <- fmt.Sprintf("%s\n", line)
				line = parts[1]
			}
		}
		close(ch)
	}()
	return ch
}

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	ch := getLinesChannel(file)
	for line := range ch {
		fmt.Printf("read: %s", line)
	}
}
