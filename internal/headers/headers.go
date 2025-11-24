package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlf_idx := bytes.Index(data, []byte(crlf))
	// need to read more
	if crlf_idx == -1 {
		return 0, false, nil
	}
	// starts with crlf, we are done
	if crlf_idx == 0 {
		return len(crlf), true, nil
	}
	data_string := string(data[:crlf_idx])
	key, val, found := strings.Cut(data_string, ":")
	if !found {
		return 0, false, fmt.Errorf("invalid header format: missing : in %s", data_string)
	}
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header format: trailing whitespace in header key %s", key)
	}
	key = strings.TrimSpace(key)
	key_has_invalid_char := strings.ContainsFunc(key, func(r rune) bool {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return false
		}
		return !strings.ContainsRune("!#$%&'*+-.^_`|~", r)
	})
	if key_has_invalid_char {
		return 0, false, fmt.Errorf("invalid header key '%s': contains invalid character", key)
	}
	key = strings.ToLower(key)
	val = strings.TrimSpace(val)

  existing_val, exists := h[key]
	if exists {
    h[key] = fmt.Sprintf("%s, %s", existing_val, val)
	} else {
		h[key] = val
	}
	return crlf_idx + len(crlf), false, nil
}
