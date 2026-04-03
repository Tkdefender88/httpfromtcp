package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Tkdefender88/httpfromtcp/internal/headers"
)

type writeState int

const (
	writeStatusLine writeState = iota
	writeHeaders
	writeBody
)

type Writer struct {
	state writeState
	w     io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state: writeStatusLine,
		w:     w,
	}
}

func (w *Writer) WriteStatus(code StatusCode) error {
	if w.state != writeStatusLine {
		return fmt.Errorf("Need to write status line first")
	}
	w.state = writeHeaders
	return WriteStatusLine(w.w, code)
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writeHeaders {
		return fmt.Errorf("Need to write headers after status line")
	}
	w.state = writeBody
	return WriteHeaders(w.w, headers)
}

func (w *Writer) WriteBody(p []byte) error {
	if w.state != writeBody {
		return fmt.Errorf("Need to write headers before body")
	}
	w.w.Write(p)
	return nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	return fmt.Fprintf(w.w, "%X\r\n%s\r\n", len(p), p)
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return fmt.Fprint(w.w, "0\r\n")
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writeBody {
		return fmt.Errorf("Not in the correct state to write trailers")
	}

	return WriteHeaders(w.w, h)
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func SetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.New()
	h.Set("Content-Length", strconv.Itoa(contentLength))
	h.Set("Connection", "close")
	return h
}

func WriteHeaders(w io.Writer, header headers.Headers) error {
	for k, v := range header {
		h := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.Write([]byte(h))
		if err != nil {
			return fmt.Errorf("error writing headers: %w", err)
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("Failed to write last header new line: %w", err)
	}
	return nil
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var response string
	switch statusCode {
	case StatusOK:
		response = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, "OK")
	case StatusBadRequest:
		response = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, "Bad Request")
	case StatusInternalServerError:
		response = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, "Internal Server Error")
	default:
		response = fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)
	}

	_, err := w.Write([]byte(response))
	if err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}
