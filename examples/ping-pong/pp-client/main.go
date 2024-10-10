package main

import (
	"time"

	"git.qowevisa.me/Qowevisa/tcpmachine/tcpclient"
)

func main() {
	conf := tcpclient.GetDefaultConfig()
	client := tcpclient.CreateClient(conf)
	go func() {
		for {
			if client.IsConnected {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		for i := 0; i < 10; i++ {
			client.Server.Write([]byte("PING\n"))
			time.Sleep(time.Second)
		}
	}()
	go client.ErrorResolver(client.ErrorsChannel)
	client.StartClient("127.0.0.1:10000")
}
