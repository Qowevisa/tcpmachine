package tcpserver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"git.qowevisa.me/Qowevisa/tcpmachine/tcpcommand"
)

type ServerConfiguration struct {
	MessageEndRune   rune
	MessageSplitRune rune
	HandleClientFunc func(client net.Conn)
	//
	ErrorResolver func(chan error)
}

func CreateHandleClientFuncFromCommands(bundle *tcpcommand.CommandBundle, conf ServerConfiguration) (func(client net.Conn), chan error) {
	clientErrors := make(chan error, 16)
	return func(client net.Conn) {
		connReader := bufio.NewReader(client)
		for {
			msg, err := connReader.ReadString(byte(conf.MessageEndRune))
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				clientErrors <- err
			}
			msgWoNl := strings.Trim(msg, string(conf.MessageEndRune))
			parts := strings.Split(msgWoNl, string(conf.MessageSplitRune))
			for _, cmd := range bundle.Commands {
				if cmd.Command == parts[0] {
					cmd.Action(parts[1:], client)
					break
				}
			}
		}
	}, clientErrors
}

type Server struct {
	HandleClientFunc func(client net.Conn)
	Exit             chan bool
	//
	ErrorsChannel chan error
	ErrorResolver func(chan error)
}

// HandleClientFunc is NOT created by this function
// see: CreateHandleClientFuncFromCommands(bundle)
func GetDefaultConfig() ServerConfiguration {
	return ServerConfiguration{
		ErrorResolver: func(c chan error) {
			for err := range c {
				log.Printf("DefConfig:Error: %v\n", err)
			}
		},
	}
}

func CreateServer(conf ServerConfiguration) *Server {
	return &Server{
		HandleClientFunc: conf.HandleClientFunc,
		Exit:             make(chan bool, 1),
		//
		ErrorsChannel: make(chan error, 8),
		ErrorResolver: conf.ErrorResolver,
	}
}

func (s *Server) StartServer(address string) error {
	if s.Exit == nil {
		return fmt.Errorf("server's Exit channel is nil. Can't start server\n")
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}
	defer listener.Close()
	if s.ErrorsChannel == nil {
		return fmt.Errorf("server's ErrorsChannel is nil. Can't start server\n")
	}
	if s.ErrorResolver == nil {
		return fmt.Errorf("server's ErrorResolver is nil. Can't start server\n")
	}
	go s.ErrorResolver(s.ErrorsChannel)
	log.Printf("Server started listening on %s\n", address)

loop:
	for {
		select {
		case <-s.Exit:
			break loop
		default:
			newCLient, err := listener.Accept()
			if err != nil {
				s.ErrorsChannel <- fmt.Errorf("listener.Accept: %w", err)
				break
			}
			go s.HandleClientFunc(newCLient)
		}
	}
	return nil
}
