package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {

	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("connection accepted")

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("read: %s\n", line)
		}

		fmt.Println("connection closed")
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
