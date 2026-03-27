package server

import (
	"fmt"
	"io"

	"github.com/Tkdefender88/httpfromtcp/internal/request"
	"github.com/Tkdefender88/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (e *HandlerError) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, e.Message)
}

func (e *HandlerError) Write(w io.Writer) {
	_ = response.WriteStatusLine(w, e.StatusCode)
	message := []byte(e.Message)
	header := response.SetDefaultHeaders(len(message))
	_ = response.WriteHeaders(w, header)
	w.Write(message)
}

type Handler func(w *response.Writer, req *request.Request)
