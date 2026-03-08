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
	assert.Equal(t, "localhost:42069", headers["host"])
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
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host:    localhost:42069   \r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["existing"] = "keep"

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

		assert.Equal(t, "keep", headers["existing"])
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, "test", headers["user-agent"])
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
		assert.Equal(t, "localhost:42069", headers["host"])
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

func TestSpecialLetters(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H@st: Localhost:42069\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestMultieLineSameKey(t *testing.T) {
	headers := NewHeaders()
	header1 := []byte("Set-Person: lane-loves-go\r\n")
	header2 := []byte("Set-Person: prime-loves-zig\r\n")
	header3 := []byte("Set-Person: tj-loves-ocaml\r\n")
	data := append(header1, header2...)
	data = append(data, header3...)

	n, done, _ := headers.Parse(header1)
	expected := "lane-loves-go"
	assert.Contains(t, headers, "set-person")
	assert.Equal(t, expected, headers["set-person"])
	assert.False(t, done)

	n, done, _ = headers.Parse(header2)
	expected = "lane-loves-go, prime-loves-zig"
	assert.Contains(t, headers, "set-person")
	assert.Equal(t, expected, headers["set-person"])
	assert.False(t, done)

	n, done, _ = headers.Parse(header3)
	expected = "lane-loves-go, prime-loves-zig, tj-loves-ocaml"
	assert.Contains(t, headers, "set-person")
	assert.Equal(t, expected, headers["set-person"])
	assert.Equal(t, len(header3), n)
	assert.False(t, done)
}
