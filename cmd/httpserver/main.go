package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

const status_400_html = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

func handle400(w *response.Writer) {
	w.WriteStatusLine(response.Status400)
	headers := response.GetDefaultHeaders(len(status_400_html))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody([]byte(status_400_html))
}

const status_500_html = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

func handle500(w *response.Writer) {
	w.WriteStatusLine(response.Status500)
	headers := response.GetDefaultHeaders(len(status_500_html))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody([]byte(status_500_html))
}

const status_200_html = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func handle200(w *response.Writer) {
	w.WriteStatusLine(response.Status200)
	headers := response.GetDefaultHeaders(len(status_200_html))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody([]byte(status_200_html))
}

func handleProxy(w *response.Writer, target string) {
	resp, err := http.Get("https://httpbin.org" + target)
	if err != nil {
		fmt.Println("failed GET to httpbin:", err)
		handle500(w)
		return
	}
	defer resp.Body.Close()
	w.WriteStatusLine(response.Status200)
	trailers := response.GetDefaultHeaders(0)
	trailers.Delete("Content-Length")
	trailers.Override("Transfer-Encoding", "chunked")
  trailers.Override("Trailer", "X-Content-Sha256, X-Content-Length")
	w.WriteHeaders(trailers)

  fullBody := []byte{}
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		fmt.Println("Read", n, "bytes")
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Println("failed reading body to buf:", err)
			}
			break
		}
		w.WriteChunkedBody(buf[:n])
    fullBody = append(fullBody, buf[:n]...)
	}
	w.WriteChunkedBodyDone()
  bodyHash := sha256.Sum256(fullBody)
  trailers = headers.NewHeaders()
  trailers.Override("X-Content-Sha256", fmt.Sprintf("%x", bodyHash))
  trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
  err = w.WriteTrailers(trailers)
  if err != nil {
    fmt.Println("failed writing trailers:", err)
  }
}

func handler(w *response.Writer, req *request.Request) {
	reqTarget := req.RequestLine.RequestTarget
	if after, ok := strings.CutPrefix(reqTarget, "/httpbin"); ok {
		httpbinTarget := after
    handleProxy(w, httpbinTarget)
		return
	}
  if reqTarget == "/yourproblem" {
		handle400(w)
    return
  }
  if reqTarget == "/myproblem" {
		handle500(w)
    return
  }
  handle200(w)
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
