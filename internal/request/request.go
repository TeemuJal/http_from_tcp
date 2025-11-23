package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type RequestState int

const (
	request_initialized RequestState = iota
	request_done
)

type Request struct {
	RequestLine RequestLine
	State       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const buffer_size = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{State: request_initialized}
	read_to_idx := 0
	buf := make([]byte, buffer_size)

	for request.State == request_initialized {
		if read_to_idx >= len(buf) {
			new_buf := make([]byte, len(buf)*2)
			copy(new_buf, buf)
			buf = new_buf
		}
		bytes_read, err := reader.Read(buf[read_to_idx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.State = request_done
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
	switch r.State {
	case request_initialized:
		data_string := string(data)
		request_line, bytes_consumed, err := parseRequestLine(data_string)
		if err != nil {
			return 0, err
		}
		if request_line != nil {
			r.RequestLine = *request_line
			r.State = 1
			return bytes_consumed, nil
		}
		return 0, nil
	case request_done:
		return 0, fmt.Errorf("error: parsing when request is done")
	default:
		return 0, fmt.Errorf("unknown request state")
	}
}

func parseRequestLine(request_string string) (*RequestLine, int, error) {
	line_split := strings.Split(request_string, "\r\n")
	if len(line_split) < 2 {
		return nil, 0, nil
	}
	request_line_split := strings.Split(line_split[0], " ")
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
	return &RequestLine{HttpVersion: http_version, RequestTarget: request_line_split[1], Method: method}, len(request_string), nil
}
