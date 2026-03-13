package requests

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParseTable(t *testing.T) {

	tests := []struct {
		name   string
		reader io.Reader
		want   RequestLine
	}{
		{
			name: "good request line",
			reader: &chunkReader{
				data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 3,
			},
			want: RequestLine{
				Method:        "GET",
				RequestTarget: "/",
				HttpVersion:   "1.1",
			},
		},
		{
			name: "good request line with path",
			reader: &chunkReader{
				data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 2,
			},
			want: RequestLine{
				Method:        "GET",
				RequestTarget: "/coffee",
				HttpVersion:   "1.1",
			},
		},
		{
			name: "good request line with post method",
			reader: &chunkReader{
				data:            "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
				numBytesPerRead: 1,
			},
			want: RequestLine{
				Method:        "POST",
				RequestTarget: "/coffee",
				HttpVersion:   "1.1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := RequestFromReader(tc.reader)
			require.NoError(t, err)
			require.NotNil(t, r)

			assert.Equal(t, tc.want.Method, r.RequestLine.Method)
			assert.Equal(t, tc.want.RequestTarget, r.RequestLine.RequestTarget)
			assert.Equal(t, tc.want.HttpVersion, r.RequestLine.HttpVersion)
		})
	}
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
