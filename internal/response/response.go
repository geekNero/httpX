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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := fmt.Sprintf("%s %d %s\r\n", HTTP, statusCode, errorCodeMap[statusCode])
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := make(headers.Headers)
	h[headers.CONTENT_LENGTH] = fmt.Sprintf("%d", contentLen)
	h[headers.CONTENT_TYPE] = "text/plain"
	h[headers.CONNECTON] = "close"
	return h
}

func WriteHeaders(w io.WriteCloser, h headers.Headers) error {
	for key, value := range h {
		fieldLine := fmt.Sprintf("%s: %s\r\n", key, value)
		w.Write([]byte(fieldLine))
	}
	w.Write([]byte("\r\n"))
	return nil
}
