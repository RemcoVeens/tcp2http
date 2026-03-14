package server

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/RemcoVeens/tcp2http/internal/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerError(t *testing.T) {
	err := HandlerError{StatusCode: 400, Message: "bad request"}
	assert.Equal(t, "bad request", err.Error())
	assert.Equal(t, 400, err.StatusCode)
}

func TestWriteStatusLine(t *testing.T) {
	tests := []struct {
		name       string
		statusCode StatusCode
		expected   string
	}{
		{"OK", OK, "HTTP/1.1 200 OK\r\n"},
		{"BadRequest", BadRequest, "HTTP/1.1 400 Bad Request\r\n"},
		{"InternalServerError", InteranlServerError, "HTTP/1.1 500 Internal Server Error\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := WriteStatusLine(buf, tt.statusCode)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestWriteStatusLineUnknown(t *testing.T) {
	buf := &bytes.Buffer{}
	err := WriteStatusLine(buf, 404)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 404\r\n", buf.String())
}

func TestGetDefaultHeaders(t *testing.T) {
	headers := GetDefaultHeaders(13)
	assert.Equal(t, "13", headers["Content-Length"])
	assert.Equal(t, "close", headers["Connection"])
	assert.Equal(t, "text/plain", headers["Content-Type"])
}

func TestGetDefaultHeadersZero(t *testing.T) {
	headers := GetDefaultHeaders(0)
	assert.Equal(t, "0", headers["Content-Length"])
}

func TestWriteHandlerError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := WriteHandlerError(buf, HandlerError{StatusCode: 400, Message: "bad request"})
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "HTTP/1.1 400 Bad Request")
	assert.Contains(t, buf.String(), "Content-Length: 11")
	assert.Contains(t, buf.String(), "bad request")
}

func TestHandle(t *testing.T) {
	tests := []struct {
		name           string
		requestTarget  string
		expectedStatus int
		expectedBody   string
	}{
		{"Default path", "/", 200, "All good, frfr\n"},
		{"Your problem", "/yourproblem", 400, "Your problem is not my problem\n"},
		{"My problem", "/myproblem", 500, "Woopsie, my bad\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			req := &request.Request{
				RequestLine: request.RequestLine{
					RequestTarget: tt.requestTarget,
				},
			}
			err := Handle(buf, req)

			if tt.expectedStatus >= 400 {
				require.Error(t, err)
				handlerErr, ok := err.(HandlerError)
				require.True(t, ok)
				assert.Equal(t, tt.expectedStatus, handlerErr.StatusCode)
				assert.Equal(t, tt.expectedBody, handlerErr.Message)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, buf.String())
			}
		})
	}
}

type mockConn struct {
	reader      *bytes.Buffer
	writer      *bytes.Buffer
	closeCalled bool
}

func (m *mockConn) Read(b []byte) (int, error) {
	return m.reader.Read(b)
}

func (m *mockConn) Write(b []byte) (int, error) {
	return m.writer.Write(b)
}

func (m *mockConn) Close() error {
	m.closeCalled = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 12345}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 54321}
}

func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestServerHandle(t *testing.T) {
	t.Run("Valid request", func(t *testing.T) {
		reqData := "GET / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
		conn := &mockConn{
			reader: bytes.NewBufferString(reqData),
			writer: &bytes.Buffer{},
		}

		server := Server{}
		server.handler = Handle

		server.handle(conn)

		assert.True(t, conn.writer.Len() > 0)
		assert.Contains(t, conn.writer.String(), "HTTP/1.1 200 OK")
		assert.Contains(t, conn.writer.String(), "All good, frfr")
	})

	t.Run("Bad request", func(t *testing.T) {
		reqData := "INVALID REQUEST\r\n"
		conn := &mockConn{
			reader: bytes.NewBufferString(reqData),
			writer: &bytes.Buffer{},
		}

		server := Server{}
		server.handler = Handle

		server.handle(conn)

		assert.True(t, conn.writer.Len() > 0)
		assert.Contains(t, conn.writer.String(), "HTTP/1.1 400 Bad Request")
	})

	t.Run("Handler error", func(t *testing.T) {
		reqData := "GET /yourproblem HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
		conn := &mockConn{
			reader: bytes.NewBufferString(reqData),
			writer: &bytes.Buffer{},
		}

		server := Server{}
		server.handler = Handle

		server.handle(conn)

		assert.True(t, conn.writer.Len() > 0)
		assert.Contains(t, conn.writer.String(), "HTTP/1.1 400 Bad Request")
		assert.Contains(t, conn.writer.String(), "Your problem is not my problem")
	})

	t.Run("Internal server error from handler", func(t *testing.T) {
		reqData := "GET /myproblem HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
		conn := &mockConn{
			reader: bytes.NewBufferString(reqData),
			writer: &bytes.Buffer{},
		}

		server := Server{}
		server.handler = Handle

		server.handle(conn)

		assert.True(t, conn.writer.Len() > 0)
		assert.Contains(t, conn.writer.String(), "HTTP/1.1 500 Internal Server Error")
		assert.Contains(t, conn.writer.String(), "Woopsie, my bad")
	})

	t.Run("Custom handler returns error", func(t *testing.T) {
		reqData := "GET / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
		conn := &mockConn{
			reader: bytes.NewBufferString(reqData),
			writer: &bytes.Buffer{},
		}

		customHandler := func(w *bytes.Buffer, r *request.Request) error {
			return io.EOF
		}

		server := Server{}
		server.handler = customHandler

		server.handle(conn)

		assert.True(t, conn.writer.Len() > 0)
		assert.Contains(t, conn.writer.String(), "HTTP/1.1 500 Internal Server Error")
	})
}
