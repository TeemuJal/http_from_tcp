package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
	"unicode"
)

type RequestState int

const (
	request_initialized RequestState = iota
	request_parsing_headers
	request_done
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const buffer_size = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{State: request_initialized, Headers: headers.NewHeaders()}
	read_to_idx := 0
	buf := make([]byte, buffer_size)

	for request.State != request_done {
		if read_to_idx >= len(buf) {
			new_buf := make([]byte, len(buf)*2)
			copy(new_buf, buf)
			buf = new_buf
		}
		bytes_read, err := reader.Read(buf[read_to_idx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.State != request_done {
					return nil, fmt.Errorf("reached EOF without request being done")
				}
				break
			}
			return nil, err
		}
		read_to_idx += bytes_read
		bytes_parsed, err := request.parse(buf[:read_to_idx])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[bytes_parsed:])
		read_to_idx -= bytes_parsed
	}
	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	total_bytes_parsed := 0
	for r.State != request_done {
		bytes_parsed, err := r.parseSingle(data[total_bytes_parsed:])
		if err != nil {
			return 0, err
		}
		if bytes_parsed == 0 {
			break
		}
		total_bytes_parsed += bytes_parsed
	}
	return total_bytes_parsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case request_initialized:
		request_line, bytes_parsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if request_line != nil {
			r.RequestLine = *request_line
			r.State = request_parsing_headers
			return bytes_parsed, nil
		}
		return 0, nil
	case request_parsing_headers:
		bytes_parsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = request_done
		}
		return bytes_parsed, nil
	case request_done:
		return 0, fmt.Errorf("error: parsing when request is done")
	default:
		return 0, fmt.Errorf("unknown request state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	crlf_idx := bytes.Index(data, []byte(crlf))
	if crlf_idx == -1 {
		return nil, 0, nil
	}
	request_line_string := string(data[:crlf_idx])
	request_line_split := strings.Split(request_line_string, " ")
	if len(request_line_split) != 3 {
		return nil, 0, errors.New("invalid number of request line parts")
	}
	method := request_line_split[0]
	for _, rune := range method {
		if !unicode.IsUpper(rune) {
			return nil, 0, errors.New("method should only contain uppercase letters")
		}
	}
	http_version_split := strings.Split(request_line_split[2], "/")
	if len(http_version_split) != 2 || http_version_split[1] != "1.1" || http_version_split[0] != "HTTP" {
		return nil, 0, errors.New("invalid HTTP version")
	}
	http_version := http_version_split[1]
	return &RequestLine{HttpVersion: http_version, RequestTarget: request_line_split[1], Method: method}, crlf_idx + len(crlf), nil
}
