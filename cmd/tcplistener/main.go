package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("failed to listen:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
  
  fmt.Println("Listening tcp traffic on port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to accept conn:", err.Error())
      break
		}
    fmt.Println("Connection has been accepted from", conn.RemoteAddr())
		linesChannel := getLinesChannel(conn)

		for line := range linesChannel {
			fmt.Println(line)
		}
    fmt.Println("connection to", conn.RemoteAddr(), "has been closed")
	}
}

func getLinesChannel(conn io.ReadCloser) <-chan string {
	linesChannel := make(chan string)

	go func() {
		defer close(linesChannel)
		currentLine := ""
		for {
			buffer := make([]byte, 8)
			n, err := conn.Read(buffer)
			if err != nil {
				if currentLine != "" {
					linesChannel <- currentLine
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Println("failed to read:", err.Error())
				break
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			if len(parts) > 1 {
				for _, part := range parts[:len(parts)-1] {
					linesChannel <- fmt.Sprintf("%s%s", currentLine, part)
					currentLine = ""
				}
			}
			currentLine += parts[len(parts)-1]
		}
	}()
	return linesChannel
}
