package main

import (
	"fmt"
	"net"
	"os"

	"github.com/Tkdefender88/httpfromtcp/internal/requests"
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

		req, err := requests.RequestFromReader(conn)
		if err != nil {
			fmt.Fprint(os.Stderr, fmt.Errorf("error reading request: %w", err))
		}

		fmt.Printf(
			"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion,
		)

		fmt.Println("connection closed")
	}
}
