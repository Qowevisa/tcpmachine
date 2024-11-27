package main

import (
	"fmt"
	"time"

	"git.qowevisa.me/qowevisa/tcpmachine/tcpclient"
)

func main() {
	client := tcpclient.CreateClient("127.0.0.1:10000")
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
			client.ErrorsChannel <- fmt.Errorf("test err")
		}
	}()
	err := client.StartClient()
	if err != nil {
		fmt.Printf("ERROR: StartClient: %v\n", err)
	}
}
