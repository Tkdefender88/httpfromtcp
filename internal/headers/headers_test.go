package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse_ValidSingleHeader(t *testing.T) {
	headers := New()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err, "expect no errors from valid header line")
	require.NotNil(t, headers, "expect headers to be non nil")
	assert.Equal(t, "localhost:42069", headers.Get("Host"), "expect to parse the host and store it the map")
	assert.Equal(t, 23, n, "parser should have parsed 23 bytes")
	assert.Equal(t, false, done, "parser should not be done")
}

func TestHeaderParse_ValidSingleHeaderExtraWhitespace(t *testing.T) {
	headers := New()
	data := []byte("Host:        localhost:42069     \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err, "expect no errors from valid header line")
	require.NotNil(t, headers, "expect headers to be non nil")
	assert.Equal(t, "localhost:42069", headers.Get("Host"), "expect to parse the host and store it the map")
	assert.Equal(t, 35, n, "parser should have parsed 23 bytes")
	assert.Equal(t, false, done, "parser should not be done")
}

func TestHeaderParse_ValidDone(t *testing.T) {
	headers := New()
	data := []byte("\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err, "expect no errors from valid header line")
	require.NotNil(t, headers, "expect headers to be non nil")
	assert.Equal(t, 0, n)
	assert.Equal(t, true, done, "parser should not be done")
}

func TestHeadersParse_DuplicateKey(t *testing.T) {
	var headers Headers = map[string]string{"host": "localhost:42069"}
	data := []byte("host: primeisgay.com\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err, "expect no errors from valid header line")
	require.NotNil(t, headers, "expect headers to be non nil")
	assert.Equal(t, "localhost:42069,primeisgay.com", headers.Get("Host"), "expect to find 'Host' in the map")
	assert.Equal(t, 22, n, "parser should have parsed 22 bytes")
	assert.Equal(t, false, done, "parser should not be done")
}

func TestHeaderParse_ValidTwoHeaders(t *testing.T) {
	var headers Headers = map[string]string{"host": "localhost:42069"}
	data := []byte("Accept: */*\r\nFoo: Bar\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err, "expect no errors from valid header line")
	require.NotNil(t, headers, "expect headers to be non nil")
	assert.Equal(t, "localhost:42069", headers.Get("Host"), "expect to find the host in the headers map")
	assert.Equal(t, "*/*", headers.Get("Accept"), "expect to find 'Accept' in the map")
	assert.Equal(t, 13, n, "parser should have parsed 23 bytes")
	assert.Equal(t, false, done, "parser should not be done")
}

func TestHeaderParse_InvalidSpacing(t *testing.T) {
	// Test: Invalid spacing header
	headers := New()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestHeaderParse_InvalidCharacter(t *testing.T) {
	// Test: Invalid spacing header
	headers := New()
	data := []byte("H©st: localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
