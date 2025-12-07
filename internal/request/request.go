package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type State int

const (
	// HTTP Methods
	Post = "POST"
	Get  = "GET"
	Put  = "PUT"

	// HTTP New Line
	CRLF = "\r\n"
)

const (
	Initialized State = iota
	Done
)

var (
	DefaultMethods = []string{
		Post,
		Get,
		Put,
	}
)

type Request struct {
	*RequestLine
	State
	rawStream string
}

// "testing"
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(line string) (*RequestLine, int, error) {
	idx := strings.Index(line, CRLF)
	if idx == -1 {
		return nil, 0, nil
	}
	line = line[:idx]

	startLineParts := strings.Split(line, " ")
	if len(startLineParts) != 3 {
		return nil, 0, errors.New("Invalid start-line format")
	}
	validMethod := false
	for _, method := range DefaultMethods {
		if method == startLineParts[0] {
			validMethod = true
		}
	}
	if !validMethod {
		return nil, 0, errors.New("Invalid Method")
	}
	requestTarget := startLineParts[1]
	httpPart := strings.Split(startLineParts[2], "/")
	if httpPart[0] != "HTTP" || httpPart[1] != "1.1" {
		return nil, 0, fmt.Errorf("Unsupported protocol: %s", startLineParts[2])
	}
	return &RequestLine{
		HttpVersion:   httpPart[1],
		RequestTarget: requestTarget,
		Method:        startLineParts[0],
	}, idx + len(CRLF), nil
}

func (r *Request) parse(data []byte) (int, error) {

	var reqLine *RequestLine
	var idx int
	var err error

	r.rawStream += string(data)
	if r.RequestLine == nil {
		reqLine, idx, err = parseRequestLine(r.rawStream)
		if err != nil {
			return 0, err
		}
	}

	if idx > 0 {
		r.rawStream = r.rawStream[idx:]
		r.RequestLine = reqLine
		r.State = Done
	}
	return idx, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{}
	for req.State == Initialized {

		reqByte := make([]byte, 32)
		n, err := reader.Read(reqByte)
		if err != nil {
			return nil, fmt.Errorf("Failed to read from reader, error: %s", err.Error())
		}
		_, err = req.parse(reqByte[:n])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse data stream, error: %s", err.Error())
		}
	}
	return req, nil
}
