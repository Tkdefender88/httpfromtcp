package requests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse_GoodRequestLine(t *testing.T) {
	r, err := RequestFromReader(
		strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"),
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestRequestLineParse_GoodRequestLineWithPath(t *testing.T) {
	r, err := RequestFromReader(
		strings.NewReader(
			"GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestRequestLineParse_GoodRequestLinePostMethod(t *testing.T) {
	r, err := RequestFromReader(
		strings.NewReader(
			"POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestRequestLineParse_FailureMissingMethod(t *testing.T) {
	_, err := RequestFromReader(
		strings.NewReader(
			"/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.Error(t, err)
}

func TestRequestLineParse_FailureOutOfOrder(t *testing.T) {
	_, err := RequestFromReader(
		strings.NewReader(
			"/coffee GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.Error(t, err)
}

func TestRequestLineParse_FailureIncompatibleVersion(t *testing.T) {
	_, err := RequestFromReader(
		strings.NewReader(
			"GET /coffee HTTP/1.2\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.Error(t, err)
}
