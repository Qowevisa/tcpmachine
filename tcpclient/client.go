package tcpclient

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
)

type ClientConfiguration struct {
	Status        uint32
	ErrorResolver func(chan error)
}

func GetDefaultConfig() *ClientConfiguration {
	return &ClientConfiguration{
		ErrorResolver: func(c chan error) {
			for err := range c {
				fmt.Printf("DefConfig:Error: %v\n", err)
			}
		},
	}
}

type ErrorResolverFunc func(errors chan error)

type Client struct {
	addr        string
	Status      uint32
	exit        chan bool
	Server      net.Conn
	IsConnected bool
	//
	Messages      chan []byte
	ErrorsChannel chan error
	ErrorResolver ErrorResolverFunc
}

func CreateClient(addr string, options ...ClientOption) *Client {
	conf := GetDefaultConfig()

	for _, opt := range options {
		opt(conf)
	}

	return &Client{
		addr:          addr,
		Messages:      make(chan []byte, 16),
		ErrorResolver: conf.ErrorResolver,
		ErrorsChannel: make(chan error, 8),
		exit:          make(chan bool, 1),
	}
}

var (
	ERROR_CLIENT_ERRRSL_NIL = errors.New("Error Resolver is nil")
	ERROR_CLIENT_ERRCHL_NIL = errors.New("Error Channel is nil")
)

func (c *Client) StartClient() error {
	if c.Status&statusBitCustomErrorHandling == 0 {
		if c.ErrorResolver == nil {
			return fmt.Errorf("Can't start client: %w", ERROR_CLIENT_ERRRSL_NIL)
		}
		if c.ErrorResolver == nil {
			return fmt.Errorf("Can't start client: %w", ERROR_CLIENT_ERRCHL_NIL)
		}
		go c.ErrorResolver(c.ErrorsChannel)
	}
	server, err := net.Dial("tcp", c.addr)
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
