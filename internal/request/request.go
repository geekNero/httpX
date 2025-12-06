package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	Post = "POST"
	Get  = "GET"
	Put  = "PUT"
)

var (
	DefaultMethods = []string{
		Post,
		Get,
		Put,
	}
)

type Request struct {
	RequestLine RequestLine
}

// "testing"
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(line string) (*RequestLine, error) {
	startLineParts := strings.Split(line, " ")
	if len(startLineParts) != 3 {
		return nil, errors.New("Invalid start-line format")
	}
	validMethod := false
	for _, method := range DefaultMethods {
		if method == startLineParts[0] {
			validMethod = true
		}
	}
	if !validMethod {
		return nil, errors.New("Invalid Method")
	}
	requestTarget := startLineParts[1]
	httpPart := strings.Split(startLineParts[2], "/")
	if httpPart[0] != "HTTP" || httpPart[1] != "1.1" {
		return nil, fmt.Errorf("Unsupported protocol: %s", startLineParts[2])
	}
	return &RequestLine{
		HttpVersion:   httpPart[1],
		RequestTarget: requestTarget,
		Method:        startLineParts[0],
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from reader, error: %s", err.Error())
	}

	lines := strings.Split(string(req), "\r\n")
	if len(lines) == 0 {
		return nil, errors.New("Invalid Request format")
	}

	reqLine, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, fmt.Errorf("Failed to parse request header: %s, error: %s", lines[0], err.Error())
	}

	return &Request{RequestLine: *reqLine}, nil
}
