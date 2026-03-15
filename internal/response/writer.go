package response

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/RemcoVeens/tcp2http/internal/headers"
)

type writerState int

const (
	writerStateExpectStatusLine writerState = iota
	writerStateExpectHeaders
	writerStateExpectBody
	writerStateBody
)

type Writer struct {
	writerState writerState
	buf         bytes.Buffer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateExpectStatusLine {
		return errors.New("WriteStatusLine called out of order")
	}
	w.writerState = writerStateExpectHeaders
	return nil
}

func (w *Writer) WriteHeaders(hdrs headers.Headers) error {
	if w.writerState != writerStateExpectHeaders {
		return errors.New("WriteHeaders called out of order")
	}
	w.writerState = writerStateExpectBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateExpectBody && w.writerState != writerStateBody {
		return 0, errors.New("WriteBody called out of order")
	}
	w.writerState = writerStateBody
	return 0, nil
}

func (w *Writer) WriteString(s string) (int, error) {
	return w.buf.WriteString(s)
}

func (w *Writer) Len() int {
	return w.buf.Len()
}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

func (w *Writer) WriteTrailers(hdrs headers.Headers) error {
	if w.writerState != writerStateBody {
		return errors.New("WriteTrailers called out of order")
	}
	for hdr := range hdrs {
		fmt.Fprintf(&w.buf, "%s: %s\r\n", hdr, hdrs[hdr])
		_, err := w.buf.Write([]byte(hdr + ": " + hdrs[hdr] + "\r\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
