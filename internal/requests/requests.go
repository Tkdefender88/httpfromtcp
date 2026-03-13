package requests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const bufferSize = 8

type parseState int

const (
	initialized parseState = iota
	done
)

type Request struct {
	RequestLine RequestLine
	parseState  parseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.parseState {
	case initialized:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.parseState = done
		return n, nil
	case done:
		return 0, fmt.Errorf("trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}

func RequestFromReader(r io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{parseState: initialized}

	for req.parseState != done {
		if readToIndex >= len(buffer) {
			newBuf := make([]byte, len(buffer)*2)
			copy(newBuf, buffer)
			buffer = newBuf
		}

		bytesRead, err := r.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.parseState != done {
					return nil, fmt.Errorf("premature end of stream")
				}
				req.parseState = done
				break
			}
			return nil, err
		}

		readToIndex += bytesRead

		bytesParsed, err := req.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buffer, buffer[bytesParsed:])
		readToIndex -= bytesParsed
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func requestLineFromString(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}

	method := parts[0]
	if err := validateMethod(method); err != nil {
		return nil, err
	}

	target := parts[1]

	httpVersion := parts[2]
	if httpVersion != "HTTP/1.1" {
		return nil, fmt.Errorf("invalid http version")
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   "1.1",
	}, nil
}

func validateMethod(method string) error {
	if strings.ToUpper(method) != method {
		return fmt.Errorf("invalid method, not all uppercase")
	}

	if strings.ContainsFunc(method, func(r rune) bool { return !unicode.IsLetter(r) }) {
		return fmt.Errorf("invalid method, contains non-letter characters, method: %s", method)
	}

	return nil
}
