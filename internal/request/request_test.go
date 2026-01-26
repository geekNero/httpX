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

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, "/", r.RequestTarget)
	assert.Equal(t, "1.1", r.HTTPVersion)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Good GET Request line with path

	reader.data = "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 1
	reader.pos = 0
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, "/coffee", r.RequestTarget)
	assert.Equal(t, "1.1", r.HTTPVersion)

	// Test: Good POST Request line with path
	reader.data = "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 4
	reader.pos = 0
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.Method)
	assert.Equal(t, "/coffee", r.RequestTarget)
	assert.Equal(t, "1.1", r.HTTPVersion)

	// Test: Invalid number of request line parts
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1 Love\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 6,
	}
	_, err = RequestFromReader(reader)

	require.Error(t, err)

	// Test: Invalid method
	reader.data = "GE /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 15
	reader.pos = 0
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid http version
	_, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.5\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Incomplete bytes
	_, err = RequestFromReader(strings.NewReader("G"))
	require.Error(t, err)

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Headers
	reader.data = "GET /coffee HTTP/1.1\r\n\r\n"
	reader.numBytesPerRead = 1
	reader.pos = 0
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, r)

	// Test: No Host Header present
	reader.data = "GET /coffee HTTP/1.1\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 1
	reader.pos = 0
	r, err = RequestFromReader(reader)
	assert.EqualError(t, err, "failed to parse headers, error: host not found in headers")
	require.Nil(t, r)

	// Test: Malformed headers: expected ':' separator missing
	reader.data = "GET /coffee HTTP/1.1\r\nUser-Agent curl/7.81.0\r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 6
	reader.pos = 0
	r, err = RequestFromReader(reader)
	assert.EqualError(t, err, "failed to parse headers, error: expected ':' separator missing")
	require.Nil(t, r)

	// Test: Malformed headers: missing header value
	reader.data = "GET /coffee HTTP/1.1\r\nUser-Agent: \r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 6
	reader.pos = 0
	r, err = RequestFromReader(reader)
	assert.EqualError(t, err, "failed to parse headers, error: missing header value")
	require.Nil(t, r)

	// Test: Malformed headers: missing header key
	reader.data = "GET /coffee HTTP/1.1\r\n: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 6
	reader.pos = 0
	r, err = RequestFromReader(reader)
	assert.EqualError(t, err, "failed to parse headers, error: missing header key")
	require.Nil(t, r)

	// Test: Malformed headers: malformed header key
	reader.data = "GET /coffee HTTP/1.1\r\n{}: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	reader.numBytesPerRead = 6
	reader.pos = 0
	r, err = RequestFromReader(reader)
	assert.EqualError(t, err, "failed to parse headers, error: malformed header key")
	require.Nil(t, r)

	// // Test: Good POST Request line with path
	// reader.data = "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	// reader.numBytesPerRead = 4
	// reader.pos = 0
	// r, err = RequestFromReader(reader)
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "POST", r.Method)
	// assert.Equal(t, "/coffee", r.RequestTarget)
	// assert.Equal(t, "1.1", r.HTTPVersion)

}
