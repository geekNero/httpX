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
	// state determines whether the RequestLine has been read or not, it's values can be Initialized and Done.
	state
	rawStream strings.Builder
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

/*
parse takes in []byte as input, and returns number of lines processed, and error encountered.
It buffers the bytes in Request.rawStream until it has sufficient bytes to process the RequestLine.
After processing the RequestLine, any extra bytes are stored in rawStream to be used ahead.
*/
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

	// This loop polls the reader until it is able to read the RequestLine
	for req.state == Initialized {

		// we poll at max Rate bytes of data in every iteration
		reqByte := make([]byte, Rate)
		n, err := reader.Read(reqByte)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("stream incomplete")
			}
			return nil, fmt.Errorf("failed to read from reader, error: %s", err.Error())
		}
		// req.parse stores all bytes in rawStream, which can be directly used in further functions
		_, err = req.parse(reqByte[:n])
		if err != nil {
			return nil, fmt.Errorf("failed to parse data stream, error: %s", err.Error())
		}
		if err == io.EOF && req.state != Done {
			return nil, fmt.Errorf("stream incomplete")
		}
	}
	// var h headers.Headers
	// for{

	// 	h.p
	// }
	return req, nil
}
