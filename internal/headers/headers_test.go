package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {

	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
func TestHeadersEdgeCases(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host:    localhost:42069   \r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["Existing"] = "keep"

		header1 := []byte("Host: localhost:42069\r\n")
		header2 := []byte("User-Agent: test\r\n")
		data := append(header1, header2...)

		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, len(header1), n)
		assert.False(t, done)

		n, done, err = headers.Parse(data[n:])
		require.NoError(t, err)
		assert.Equal(t, len(header2), n)
		assert.False(t, done)

		assert.Equal(t, "keep", headers["Existing"])
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, "test", headers["User-Agent"])
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		header := []byte("Host: localhost:42069\r\n")
		n, done, err := headers.Parse(header)
		require.NoError(t, err)
		assert.Equal(t, len(header), n)
		assert.False(t, done)

		data := []byte("\r\n")
		n, done, err = headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.True(t, done)
		assert.Equal(t, "localhost:42069", headers["Host"])
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})
}
