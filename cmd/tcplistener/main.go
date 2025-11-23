package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
	"os"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Failed to listen:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
  
  fmt.Println("Listening tcp traffic on port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept conn:", err.Error())
      break
		}
    fmt.Println("Connection has been accepted from", conn.RemoteAddr())
    request, err := request.RequestFromReader(conn)
    if err != nil {
      log.Fatalln("Failed to read from connection:", err.Error())
    }
    fmt.Println("Request line:")
    fmt.Println("- Method:", request.RequestLine.Method)
    fmt.Println("- Target:", request.RequestLine.RequestTarget)
    fmt.Println("- Version:", request.RequestLine.HttpVersion)
    fmt.Println("Connection to", conn.RemoteAddr(), "has been closed")
	}
}
