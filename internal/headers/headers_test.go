package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	val, err := headers.Get("Host")
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", val)
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Valid double headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n Hell2: kennys\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "kennys", headers["hell2"])
	assert.Equal(t, 39, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: valid spacing header
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, 39, n)
	assert.True(t, done)

	// Test: Incomplete Value
	headers = NewHeaders()
	data = []byte("       Host:\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid special character in header
	headers = NewHeaders()
	data = []byte("       H©st: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

}
