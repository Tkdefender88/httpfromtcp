package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

type Server struct {
	Port int

	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error opening listener: %w", err)
	}

	server := &Server{
		Port:     port,
		listener: ln,
		closed:   atomic.Bool{},
	}

	go server.listen()

	return server, nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.closed.Load() {
			return
		}

		if err != nil {
			fmt.Printf("error accepting connnection %v\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!\n"))
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}
