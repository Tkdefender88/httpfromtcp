package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/Tkdefender88/httpfromtcp/internal/headers"
)

const bufferSize = 8

type parseState int

const (
	requestStateInitialized parseState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	parseState  parseState
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.parseState {
	case requestStateInitialized:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("error parsing request line: %w", err)
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.parseState = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, finished, err := r.Headers.Parse(data)
		if err != nil {
			return 0, fmt.Errorf("error parsing header: %w", err)
		}
		if finished {
			r.parseState = requestStateParsingBody
		}
		if n == 0 {
			return 0, nil
		}
		return n, nil
	case requestStateParsingBody:
		contentLength := r.Headers.Get("content-length")
		if contentLength == "" {
			r.parseState = requestStateDone
			return len(data), nil
		}

		reportedLength, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, fmt.Errorf("invalid content length: %q, could not parse: %w", contentLength, err)
		}

		fmt.Println(string(data))
		r.Body = append(r.Body, data...)

		if len(r.Body) > reportedLength {
			return 0, fmt.Errorf(
				"reported content length is incorrect, body is %d, reported Length is %d",
				len(r.Body),
				reportedLength,
			)
		}

		if len(r.Body) == reportedLength {
			r.parseState = requestStateDone
		}

		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytes := 0

	for r.parseState != requestStateDone {
		n, err := r.parseSingle(data[totalBytes:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytes += n
	}

	return totalBytes, nil
}

func RequestFromReader(r io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		parseState: requestStateInitialized,
		Headers:    make(headers.Headers),
	}

	for req.parseState != requestStateDone {
		if readToIndex >= len(buffer) {
			newBuf := make([]byte, len(buffer)*2)
			copy(newBuf, buffer)
			buffer = newBuf
		}

		bytesRead, readErr := r.Read(buffer[readToIndex:])
		readToIndex += bytesRead

		bytesParsed, parseErr := req.parse(buffer[:readToIndex])
		if parseErr != nil {
			return nil, fmt.Errorf("error parsing request: %w", parseErr)
		}
		copy(buffer, buffer[bytesParsed:])
		readToIndex -= bytesParsed

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				if req.parseState != requestStateDone {
					return nil, fmt.Errorf("premature end of stream")
				}
				req.parseState = requestStateDone
				break
			}
			return nil, readErr
		}
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
