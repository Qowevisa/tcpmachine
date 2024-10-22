package tcpclient

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"git.qowevisa.me/Qowevisa/tcpmachine/tcpcommand"
)

type ClientConfiguration struct {
	MessageEndRune   rune
	MessageSplitRune rune
	//
	Status        uint32
	ErrorResolver func(chan error)
	//
	ServerHandlerFunc func(server net.Conn)
	//
}

func GetDefaultConfig() *ClientConfiguration {
	return &ClientConfiguration{
		MessageEndRune:   '\n',
		MessageSplitRune: ' ',
		ErrorResolver: func(c chan error) {
			for err := range c {
				fmt.Printf("DefConfig:Error: %v\n", err)
			}
		},
	}
}

type ErrorResolverFunc func(errors chan error)

type Client struct {
	MessageEndRune   rune
	MessageSplitRune rune
	//
	addr        string
	Status      uint32
	exit        chan bool
	Server      net.Conn
	IsConnected bool
	//
	ServerHandlerFunc func(server net.Conn)
	//
	ErrorsChannel chan error
	ErrorResolver ErrorResolverFunc
	//
	Commands []tcpcommand.Command
}

func CreateClient(addr string, options ...ClientOption) *Client {
	conf := GetDefaultConfig()

	for _, opt := range options {
		opt(conf)
	}
	c := &Client{
		MessageEndRune:    conf.MessageEndRune,
		MessageSplitRune:  conf.MessageSplitRune,
		addr:              addr,
		ErrorResolver:     conf.ErrorResolver,
		ErrorsChannel:     make(chan error, 8),
		exit:              make(chan bool, 1),
		ServerHandlerFunc: conf.ServerHandlerFunc,
	}
	if c.ServerHandlerFunc == nil {
		c.ServerHandlerFunc = GetDefaultServerHandlerFunc(c)
	}

	return c
}

var (
	ERROR_CLIENT_ERRRSL_NIL = errors.New("Error Resolver is nil")
	ERROR_CLIENT_ERRCHL_NIL = errors.New("Error Channel is nil")
	ERROR_CLIENT_SRVHND_NIL = errors.New("Server Handler Func is nil")
)

func GetDefaultServerHandlerFunc(c *Client) func(server net.Conn) {
	return func(server net.Conn) {
		serverReader := bufio.NewReader(server)
		for {
			select {
			case <-c.exit:
				return
			default:
				rawMsg, err := serverReader.ReadString(byte(c.MessageEndRune))
				if err != nil {
					if errors.Is(err, io.EOF) {
						c.exit <- true
						break
					}
					c.ErrorsChannel <- fmt.Errorf("serverReader.ReadString: %w", err)
				}
				fmt.Printf("Server send us a message: %s", rawMsg)
				msg := strings.TrimRight(rawMsg, string(c.MessageEndRune))
				parts := strings.Split(msg, string(c.MessageSplitRune))
				// ???
				if len(parts) == 0 {
					continue
				}
				cmd := parts[0]
				found := false
				for _, _cmd := range c.Commands {
					if cmd == _cmd.Command {
						found = true
						_cmd.Action(parts[1:], server)
						break
					}
				}
				if !found {
					fmt.Printf("Command %s was not handled\n", cmd)
				}
			}
		}
	}
}

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
	if c.ServerHandlerFunc == nil {
		return fmt.Errorf("Can't start client: %w", ERROR_CLIENT_SRVHND_NIL)
	}
	server, err := net.Dial("tcp", c.addr)
	if err != nil {
		return fmt.Errorf("net.Dial: %w", err)
	}
	c.IsConnected = true
	c.Server = server
	c.ServerHandlerFunc(server)
	return nil
}

var commandDuplicateError = errors.New("Command already exists in server")
var commandNotHandledError = errors.New("Command was not handled")

func (c *Client) On(command string, action func(args []string, server net.Conn)) error {
	for _, cmd := range c.Commands {
		if cmd.Command == command {
			return fmt.Errorf("Failed addding command %s: %w ", command, commandDuplicateError)
		}
	}
	c.Commands = append(c.Commands, tcpcommand.Command{
		Command: command,
		Action:  action,
	})
	return nil
}
