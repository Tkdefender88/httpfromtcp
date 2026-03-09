package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	msgFile, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lines := getLinesChannel(msgFile)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(r io.ReadCloser) <-chan string {
	strChan := make(chan string)

	go func() {
		defer close(strChan)
		buf := make([]byte, 8)
		lineContents := ""
		for {
			n, err := r.Read(buf)
			if err != nil {
				if lineContents != "" {
					strChan <- lineContents
				}
				if errors.Is(err, io.EOF) {
					r.Close()
					break
				}
				break
			}
			str := string(buf[:n])
			parts := strings.Split(str, "\n")
			for _, part := range parts[:len(parts)-1] {
				lineContents += part
				strChan <- lineContents
				lineContents = ""
			}
			lineContents += parts[len(parts)-1]
		}
	}()

	return strChan
}
