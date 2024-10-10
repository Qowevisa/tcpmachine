package main

import (
	"fmt"
	"net"
	"time"

	"git.qowevisa.me/Qowevisa/tcpmachine/tcpcommand"
	"git.qowevisa.me/Qowevisa/tcpmachine/tcpserver"
)

func main() {
	bundle, err := tcpcommand.CreateCommandBundle([]tcpcommand.Command{
		{
			Command: "PING",
			Action: func(s []string, client net.Conn) {
				fmt.Printf("todo..\n")
				client.Write([]byte("PONG\n"))
			},
		},
	})
	if err != nil {
		panic(err)
	}
	conf := tcpserver.GetDefaultConfig()
	handler, errorChannel := tcpserver.CreateHandleClientFuncFromCommands(bundle, conf)
	go func() {
		for err := range errorChannel {
			fmt.Printf("Error:1: %v\n", err)
		}
	}()
	conf.HandleClientFunc = handler
	server := tcpserver.CreateServer(conf)
	go func() {
		time.Sleep(time.Minute)
		server.Exit <- true
	}()
	server.StartServer("127.0.0.1:10000")
}
