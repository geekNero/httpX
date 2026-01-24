package request

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"basic_protocol/internal/common"
)

type state int

const (
	// HTTP Methods
	Post = "POST"
	Get  = "GET"
	Put  = "PUT"

	// Input Stream Rate
	Rate = 32
)

const (
	Initialized state = iota
	Done
)

var DefaultMethods = []string{
	Post,
	Get,
	Put,
}

type Request struct {
	*RequestLine
	state
	rawStream strings.Builder
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(line string) (*RequestLine, int, error) {
	idx := strings.Index(line, common.CRLF)
	if idx == -1 {
		return nil, 0, nil
	}
	line = line[:idx]

	startLineParts := strings.Split(line, " ")
	if len(startLineParts) != 3 {
		return nil, 0, errors.New("invalid start-line format")
	}
	validMethod := false
	for _, method := range DefaultMethods {
		if method == startLineParts[0] {
			validMethod = true
		}
	}
	if !validMethod {
		return nil, 0, errors.New("invalid Method")
	}
	requestTarget := startLineParts[1]
	httpPart := strings.Split(startLineParts[2], "/")
	if httpPart[0] != "HTTP" || httpPart[1] != "1.1" {
		return nil, 0, fmt.Errorf("unsupported protocol: %s", startLineParts[2])
	}
	return &RequestLine{
		HTTPVersion:   httpPart[1],
		RequestTarget: requestTarget,
		Method:        startLineParts[0],
	}, idx + len(common.CRLF), nil
}

func (r *Request) parse(data []byte) (int, error) {
	var reqLine *RequestLine
	var idx int
	var err error

	r.rawStream.Write(data)
	rawStream := r.rawStream.String()
	if r.RequestLine == nil {
		reqLine, idx, err = parseRequestLine(rawStream)
		if err != nil {
			return 0, err
		}
	}

	if idx > 0 {
		r.rawStream.Reset()
		r.rawStream.WriteString(rawStream[idx:])
		r.RequestLine = reqLine
		r.state = Done
	}
	return idx, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{}
	for req.state == Initialized {

		reqByte := make([]byte, Rate)
		n, err := reader.Read(reqByte)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read from reader, error: %s", err.Error())
		}
		_, err = req.parse(reqByte[:n])
		if err != nil {
			return nil, fmt.Errorf("failed to parse data stream, error: %s", err.Error())
		}
		if err == io.EOF && req.state != Done {
			return nil, fmt.Errorf("stream incomplete")
		}
	}
	return req, nil
}
