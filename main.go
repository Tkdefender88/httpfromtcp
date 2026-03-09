package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	msgFile, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer msgFile.Close()

	buf := make([]byte, 8)
	for {
		n, err := msgFile.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Fprintf(os.Stderr, "error reading: %s\n", err)
			os.Exit(1)
		}
		str := string(buf[:n])
		fmt.Printf("read: %s\n", str)
	}
}
