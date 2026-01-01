package headers

import (
	"basic_protocol/internal/common"
	"bytes"
	"fmt"
	"strings"
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
			return n, true, nil
		}

		key, value, err := parseFieldLine(parsableString)
		if err != nil {
			return 0, false, err
		}

		h[key] = value

		if n == len(data) {
			return n, false, nil
		}
	}
}
