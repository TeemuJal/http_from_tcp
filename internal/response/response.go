package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	Status200 StatusCode = 200
	Status400 StatusCode = 400
	Status500 StatusCode = 500
)

type writerState int

const (
	writerInitialized writerState = iota
  writerStatusLineWritten
  writerHeadersWritten
  writerBodyWritten
)

type Writer struct {
  writer        io.Writer
  writerState   writerState
}

func NewWriter(w io.Writer) *Writer {
  return &Writer{ writer: w, writerState: writerInitialized }
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
  if w.writerState != writerInitialized {
    return fmt.Errorf("error: writing status line in state %d", w.writerState)
  }
  reasonPhrase := ""
	switch statusCode {
	case Status200:
    reasonPhrase = "OK"
	case Status400:
    reasonPhrase = "Bad Request"
	case Status500:
    reasonPhrase = "Internal Server Error"
  default:
    reasonPhrase = ""
	}
  _, err := fmt.Fprintf(w.writer, "HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
  w.writerState = writerStatusLineWritten
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = fmt.Sprintf("%d", contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"
	return headers
}

func (w *Writer) writeHeaders(headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}
  _, err := w.writer.Write([]byte("\r\n"))
  return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
  if w.writerState != writerStatusLineWritten {
    return fmt.Errorf("error: writing headers in state %d", w.writerState)
  }
  err := w.writeHeaders(headers)
  w.writerState = writerHeadersWritten
  return err
}

func (w *Writer) WriteTrailers(trailers headers.Headers) error {
  if w.writerState != writerBodyWritten {
    return fmt.Errorf("error: writing trailers in state %d", w.writerState)
  }
  err := w.writeHeaders(trailers)
  if err != nil {
    return err
  }
  _, err = w.writer.Write([]byte("\r\n"))
  return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
  if w.writerState != writerHeadersWritten {
    return 0, fmt.Errorf("error: writing body in state %d", w.writerState)
  }
  w.writerState = writerBodyWritten
  return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
  if w.writerState != writerHeadersWritten {
    return 0, fmt.Errorf("error: writing body in state %d", w.writerState)
  }
  chunkSizeHex := fmt.Sprintf("%X", len(p))
  nTotal := 0
  n, err := w.writer.Write([]byte(chunkSizeHex + "\r\n"))
  if err != nil {
    return nTotal, err
  }
  nTotal += n

  n, err = w.writer.Write(p)
  if err != nil {
    return nTotal, err
  }
  nTotal += n

  n, err = w.writer.Write([]byte("\r\n"))
  if err != nil {
    return nTotal, err
  }
  nTotal += n
  return nTotal, nil
}
func (w *Writer) WriteChunkedBodyDone() (int, error) {
  if w.writerState != writerHeadersWritten {
    return 0, fmt.Errorf("error: writing body in state %d", w.writerState)
  }
  n, err := w.writer.Write([]byte("0\r\n"))
  w.writerState = writerBodyWritten
  return n, err
}
