package request

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

func TestRequestParser_ParseBody_ValidBody(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		reader       io.Reader
		expectedBody string
	}{
		{
			name: "Standard Body",
			reader: &chunkReader{
				data: "POST /submit HTTP/1.1\r\n" +
					"Host: localhost:42069\r\n" +
					"Content-Length: 13\r\n" +
					"\r\n" +
					"hello world!\n",
				numBytesPerRead: 3,
			},
			expectedBody: "hello world!\n",
		},
		{
			name: "Empty Body, Content Length reported 0",
			reader: &chunkReader{
				data: "POST /submit HTTP/1.1\r\n" +
					"Host: localhost:42069\r\n" +
					"Content-Length: 0\r\n" +
					"\r\n",
				numBytesPerRead: 3,
			},
			expectedBody: "",
		},
		{
			name: "Empty Body, Content Length omitted",
			reader: &chunkReader{
				data: "POST /submit HTTP/1.1\r\n" +
					"Host: localhost:42069\r\n" +
					"\r\n",
				numBytesPerRead: 3,
			},
			expectedBody: "",
		},
		{
			name: "Body Exists, Content Length omitted",
			reader: &chunkReader{
				data: "POST /submit HTTP/1.1\r\n" +
					"Host: localhost:42069\r\n" +
					"\r\n" +
					"hello world!\n",
				numBytesPerRead: 3,
			},
			expectedBody: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := RequestFromReader(tc.reader)
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, tc.expectedBody, string(r.Body), "Did not get the expected body in the request")
		})
	}
}

func TestRequestParser_ParseBody_BodyTooShort(t *testing.T) {
	// Test: Body shorter than reported content length
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestRequestParser_ParseHeaders(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))
}

func TestRequestParser_ParseHeaders_DuplicateHeaders(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nHost: litfam420.com\r\n\r\n",
		numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	assert.Equal(t, "localhost:42069,litfam420.com", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))
}

func TestRequestParser_ParseHeadersMalformed(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nhost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}

	_, err := RequestFromReader(reader)
	require.Error(t, err)
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
