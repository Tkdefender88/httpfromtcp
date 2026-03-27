package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/Tkdefender88/httpfromtcp/internal/request"
	"github.com/Tkdefender88/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error opening listener: %w", err)
	}

	server := &Server{
		listener: ln,
		closed:   atomic.Bool{},
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			fmt.Printf("error accepting connnection %v\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	rw := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		body := []byte("Error occurred reading the request")
		rw.WriteStatus(response.StatusBadRequest)
		rw.WriteHeaders(response.SetDefaultHeaders(len(body)))
		rw.WriteBody(body)
		return
	}

	s.handler(rw, req)
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
