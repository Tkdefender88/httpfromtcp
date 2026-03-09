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
	defer msgFile.Close()

	buf := make([]byte, 8)
	line := ""
	for {
		n, err := msgFile.Read(buf)
		if err != nil {
			if line != "" {
				fmt.Printf("read: %s\n", line)
				line = ""
			}
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Fprintf(os.Stderr, "error reading: %s\n", err)
			os.Exit(1)
		}
		str := string(buf[:n])
		parts := strings.Split(str, "\n")
		for _, part := range parts[:len(parts)-1] {
			line += part
			fmt.Printf("read: %s\n", line)
			line = ""
		}
		line += parts[len(parts)-1]
	}
}
