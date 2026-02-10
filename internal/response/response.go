package response

import (
	"fmt"
	"io"
	"strconv"

	"basic_protocol/internal/common"
	"basic_protocol/internal/headers"
)

type StatusCode int

const (
	// HTTP Error Codes
	StatusBadRequest          StatusCode = 400
	StatusOK                  StatusCode = 200
	StatusInternalServerError StatusCode = 500

	// Protocol
	HTTP = "HTTP/1.1"
)

var errorCodeMap = map[StatusCode]string{
	StatusBadRequest:          "Bad Request",
	StatusOK:                  "OK",
	StatusInternalServerError: "Internal Server Error",
}

type WriterState string

const (
	StatusLine WriterState = "StatusLine"
	Headers    WriterState = "Headers"
	Body       WriterState = "Body"
	Done       WriterState = "Done"
	Errored    WriterState = "Errored"
)

type Writer struct {
	WriterState
	writer io.Writer
}

func NewResponseWriter(conn io.Writer) *Writer {
	return &Writer{
		WriterState: StatusLine,
		writer:      conn,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterState != StatusLine {
		return fmt.Errorf("writer state expected %s, got: %s", StatusLine, w.WriterState)
	}

	statusLine := fmt.Sprintf("%s %d %s\r\n", HTTP, statusCode, errorCodeMap[statusCode])
	_, err := w.writer.Write([]byte(statusLine))
	if err != nil {
		w.WriterState = Errored
		return err
	}

	w.WriterState = Headers
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := make(headers.Headers)
	if contentLen > 0 {
		h[headers.CONTENT_LENGTH] = fmt.Sprintf("%d", contentLen)
	}
	h[headers.CONTENT_TYPE] = "text/plain"
	h[headers.CONNECTON] = "close"
	return h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.WriterState != Headers {
		return fmt.Errorf("writer state expected %s, got: %s", Headers, w.WriterState)
	}

	for key, value := range h {
		fieldLine := fmt.Sprintf("%s: %s\r\n", key, value)
		w.writer.Write([]byte(fieldLine))
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		w.WriterState = Errored
		return err
	}
	w.WriterState = Body
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriterState != Body {
		return 0, fmt.Errorf("writer state expected %s, got: %s", Body, w.WriterState)
	}
	n, err := w.writer.Write(p)
	if err != nil {
		w.WriterState = Errored
		return n, err
	}
	w.WriterState = Done
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.WriterState != Body {
		return 0, fmt.Errorf("writer state expected %s, got: %s", Body, w.WriterState)
	}
	// Format the size header
	head := strconv.FormatInt(int64(len(p)), 16) + common.CRLF

	// Use io.MultiWriter or just write sequentially
	// The "errors" can be handled concisely:
	if _, err := w.writer.Write([]byte(head)); err != nil {
		return 0, err
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}
	if _, err := w.writer.Write([]byte(common.CRLF)); err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.WriterState != Body {
		return 0, fmt.Errorf("writer state expected %s, got: %s", Body, w.WriterState)
	}

	lastChunk := []byte("0" + common.CRLF + common.CRLF)
	return w.writer.Write(lastChunk)
}
