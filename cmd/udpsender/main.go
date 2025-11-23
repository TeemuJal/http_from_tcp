package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
  raddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
  if err != nil {
    log.Fatalln("error:", err.Error())
  }
  conn, err := net.DialUDP("udp", nil, raddr)
  if err != nil {
    log.Fatalln("error:", err.Error())
  }
  defer conn.Close()
  reader := bufio.NewReader(os.Stdin)

  for {
    fmt.Print(">")
    string, err := reader.ReadString('\n')
    if err != nil {
      log.Println("failed to read:", err.Error())
      continue
    }
    _, writeErr := conn.Write([]byte(string))
    if writeErr != nil {
      log.Println("failed to write:", writeErr.Error())
    }
  }
}
