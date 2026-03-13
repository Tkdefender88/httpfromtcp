package requests

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(r io.Reader) (Request, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Request{}, err
	}

	lines := strings.Split(string(data), "\r\n")
	if len(lines) < 1 {
		return Request{}, fmt.Errorf("invalid request")
	}

	rl, err := parseRequestLine(lines[0])
	if err != nil {
		return Request{}, err
	}

	return Request{
		RequestLine: rl,
	}, nil
}

func parseRequestLine(line string) (RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, fmt.Errorf("invalid request line")
	}

	method := parts[0]
	if err := validateMethod(method); err != nil {
		return RequestLine{}, err
	}

	target := parts[1]

	httpVersion := parts[2]
	if httpVersion != "HTTP/1.1" {
		return RequestLine{}, fmt.Errorf("invalid http version")
	}

	return RequestLine{
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
