package tcpclient

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
)

type ClientConfiguration struct {
	ErrorResolver func(chan error)
}

func GetDefaultConfig() ClientConfiguration {
	return ClientConfiguration{
		ErrorResolver: func(c chan error) {
			for err := range c {
				fmt.Printf("DefConfig:Error: %v\n", err)
			}
		},
	}
}

type Client struct {
	exit        chan bool
	Server      net.Conn
	IsConnected bool
	//
	Messages      chan []byte
	ErrorsChannel chan error
	ErrorResolver func(chan error)
}

func CreateClient(conf ClientConfiguration) *Client {
	return &Client{
		Messages:      make(chan []byte, 16),
		ErrorResolver: conf.ErrorResolver,
		ErrorsChannel: make(chan error, 8),
		exit:          make(chan bool, 1),
	}
}

func (c *Client) StartClient(addr string) error {
	server, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("net.Dial: %w", err)
	}
	serverReader := bufio.NewReader(server)
	c.IsConnected = true
	c.Server = server
loop:
	for {
		select {
		case <-c.exit:
			break loop
		default:
			msg, err := serverReader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					c.exit <- true
					break
				}
				c.ErrorsChannel <- fmt.Errorf("serverReader.ReadString: %w", err)
			}
			fmt.Printf("Server send us a message: %s", msg)
			c.Messages <- []byte(msg)
		}
	}
	return nil
}
