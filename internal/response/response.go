package response

import (
	"basic_protocol/internal/headers"
	"fmt"
	"io"
)

type StatusCode int

const (
	// HTTP Error Codes
	BadRequest          StatusCode = 400
	OK                  StatusCode = 200
	InternalServerError StatusCode = 500

	// Protocol
	HTTP = "HTTP/1.1"
)

var (
	errorCodeMap = map[StatusCode]string{
		BadRequest:          "Bad Request",
		OK:                  "OK",
		InternalServerError: "Internal Server Error",
	}
)

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
	h[headers.CONTENT_LENGTH] = fmt.Sprintf("%d", contentLen)
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
