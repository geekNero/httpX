// Package headers handles HTTP request headers
package headers

import (
	"bytes"
	"fmt"
	"strings"

	"basic_protocol/internal/common"
)

const (
	HOST           = "host"
	CONTENT_LENGTH = "content-length"
	CONNECTON      = "connection"
	CONTENT_TYPE   = "content-type"
)

type Headers map[string]string

func NewHeaders() Headers {
	mp := Headers(make(map[string]string))
	return mp
}

func (h Headers) Get(key string) (string, error) {
	key = strings.ToLower(key)
	if !isValidHeaderKey(key) {
		return "", fmt.Errorf("malformed key")
	}
	return h[key], nil
}

func isAlphaNumeric(c byte) bool {
	// Check if the byte value falls within the range of alphanumeric characters
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isAllowedSpecialCharacter(c byte) bool {
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

func isValidHeaderKey(s string) bool {
	for _, char := range s {
		if !(isAlphaNumeric(byte(char)) || isAllowedSpecialCharacter(byte(char))) {
			return false
		}
	}
	return true
}

// parseFieldLine parses an entire header line, returns key (string), value (string)
// and error encountered (error)
func parseFieldLine(line string) (string, string, error) {
	keyEnd := strings.Index(line, ":")
	if keyEnd == -1 {
		return "", "", fmt.Errorf("expected ':' separator missing")
	}

	if keyEnd == len(line)-1 {
		return "", "", fmt.Errorf("missing header value")
	}

	key := line[:keyEnd]
	if key == "" {
		return "", "", fmt.Errorf("missing header key")
	}

	if !isValidHeaderKey(key) {
		return "", "", fmt.Errorf("malformed header key")
	}

	key = strings.ToLower(key)

	value := strings.Trim(line[keyEnd+1:], " ")
	if value == "" {
		return "", "", fmt.Errorf("missing header value")
	}

	return key, value, nil
}

// Set adds the header key with the given value if the key isn't already present, else
// it appends the value.
func (h Headers) Set(key, value string) error {
	if !isValidHeaderKey(key) {
		return fmt.Errorf("header key contains disallowed characters")
	}
	key = strings.ToLower(key)
	if h[key] == "" {
		h[key] = value
	} else {
		h[key] = strings.Join([]string{h[key], value}, ", ")
	}
	return nil
}

// Override unlike Set updates the value of the key to the parameter value if the key is already present,
// else it adds the key.
func (h Headers) Override(key, value string) error {
	if !isValidHeaderKey(key) {
		return fmt.Errorf("header key contains disallowed characters")
	}
	key = strings.ToLower(key)
	h[key] = value
	return nil
}

// Parse processess header data in request. It takes in a byte array and returns
// number of bytes consumed (int), header section ended (bool) and error encountered (error)
func (h Headers) Parse(data []byte) (int, bool, error) {
	breakLine := []byte(common.CRLF)
	var n int

	for {
		idx := bytes.Index(data[n:], breakLine)
		if idx == -1 {
			return n, false, nil
		}

		parsableString := string(data[n : n+idx])
		parsableString = strings.Trim(parsableString, " ")
		n += idx + len(common.CRLF)

		if parsableString == "" {
			hostValue, _ := h.Get(HOST)
			if hostValue == "" {
				return n, false, fmt.Errorf("host not found in headers")
			}
			return n, true, nil
		}

		key, value, err := parseFieldLine(parsableString)
		if err != nil {
			return 0, false, err
		}

		h.Set(key, value)

		if n == len(data) {
			return n, false, nil
		}
	}
}
