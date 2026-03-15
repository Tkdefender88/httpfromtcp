package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func New() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 0, true, nil
	}

	fieldLine := string(data[:idx])
	parts := strings.SplitN(fieldLine, ":", 2)

	key := parts[0]
	if err = validate(key); err != nil {
		return 0, false, fmt.Errorf("error parsing: %w\n", err)
	}

	key = strings.TrimSpace(key)

	value := parts[1]
	value = strings.TrimSpace(value)

	h.Set(key, value)

	return idx + 2, false, nil
}

func validate(key string) error {
	if cut := strings.TrimRight(key, " "); cut != key {
		fmt.Printf("cut: %q\n", cut)
		return fmt.Errorf("invalid header line, spaces between colon and header")
	}

	if strings.ContainsFunc(key, func(r rune) bool {
		return !strings.ContainsRune(
			"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&'*+-.^_`|~", r,
		)
	}) {
		return fmt.Errorf("Invalid characters in header key")
	}

	return nil
}

func (h Headers) Get(key string) (value string) {
	key = strings.ToLower(key)
	return h[key]
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if val, ok := h[key]; ok {
		value = fmt.Sprintf("%s,%s", val, value)
		h[key] = value
		return
	}
	h[key] = value
}
