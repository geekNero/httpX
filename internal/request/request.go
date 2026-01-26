package request

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"basic_protocol/internal/common"
	"basic_protocol/internal/headers"
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
	// state determines whether the RequestLine has been read or not, it's values can be Initialized and Done.
	state
	rawStream []byte
	headers.Headers
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(line string) (*RequestLine, int, error) {
	// Waiting until we hit the CRLF token
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

func (r *Request) wrapperRequestLine(data []byte) (int, bool, error) {
	reqLine, idx, err := parseRequestLine(string(data))
	if err != nil {
		return 0, false, err
	}
	if idx > 0 {
		r.RequestLine = reqLine
		return idx, true, nil
	} else {
		return idx, false, nil
	}
}

/*
parse takes in []byte as input, and returns number of lines processed, and error encountered.
It buffers the bytes in Request.rawStream until it has sufficient bytes to process the RequestLine.
After processing the RequestLine, any extra bytes are stored in rawStream to be used ahead.
*/
func (r *Request) parse(data []byte, caller func(data []byte) (int, bool, error)) (int, bool, error) {

	r.rawStream = append(r.rawStream, data...)

	n, done, err := caller(r.rawStream)
	r.rawStream = r.rawStream[n:]
	return n, done, err
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{}
	var n int
	var done bool
	var err error
	reqByte := make([]byte, Rate)
	// This loop polls the reader until it is able to read the RequestLine
	for !done {

		// we poll at max Rate bytes of data in every iteration
		n, err := reader.Read(reqByte)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("stream incomplete")
			}
			return nil, fmt.Errorf("failed to read for requestLine, error: %s", err.Error())
		}
		// req.parse stores all bytes in rawStream, which can be directly used in further functions
		_, done, err = req.parse(reqByte[:n], req.wrapperRequestLine)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data stream, error: %s", err.Error())
		}
	}

	h := make(headers.Headers)
	// setting n to 0, as we want to take a pass with the existing leftover rawData
	// this is also why we are parsing first and reading second
	n = 0
	done = false
	reqByte = make([]byte, Rate)
	// Iterating for headers
	for {
		_, done, err = req.parse(reqByte[:n], h.Parse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse headers, error: %s", err.Error())
		}
		if done {
			break
		}
		n, err = reader.Read(reqByte)
		if err != nil {
			return nil, fmt.Errorf("failed to read data for headers, error: %s", err.Error())
		}
	}
	req.state = Done
	req.Headers = h
	return req, nil
}
