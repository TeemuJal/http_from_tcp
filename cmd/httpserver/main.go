package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
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

const status_500_html = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const status_200_html = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func handler(w *response.Writer, req *request.Request) {
  switch req.RequestLine.RequestTarget {
  case "/yourproblem":
    w.WriteStatusLine(response.Status400)
    headers := response.GetDefaultHeaders(len(status_400_html))
    headers.Override("Content-Type", "text/html")
    w.WriteHeaders(headers)
    w.WriteBody([]byte(status_400_html))
  case "/myproblem":
    w.WriteStatusLine(response.Status500)
    headers := response.GetDefaultHeaders(len(status_500_html))
    headers.Override("Content-Type", "text/html")
    w.WriteHeaders(headers)
    w.WriteBody([]byte(status_500_html))
  default:
    w.WriteStatusLine(response.Status200)
    headers := response.GetDefaultHeaders(len(status_200_html))
    headers.Override("Content-Type", "text/html")
    w.WriteHeaders(headers)
    w.WriteBody([]byte(status_200_html))
  }
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
