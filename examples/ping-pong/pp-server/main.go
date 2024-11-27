package main

import (
	"fmt"
	"net"
	"time"

	"git.qowevisa.me/qowevisa/tcpmachine/tcpserver"
)

func main() {
	server := tcpserver.CreateServer("127.0.0.1:10000")
	server.On("PING", func(args []string, client net.Conn) {
		client.Write([]byte("PONG\n"))
	})
	time.Sleep(time.Second)
	err := server.StartServer()
	if err != nil {
		fmt.Printf("Error: server.StartServer: %v\n", err)
	}
}
