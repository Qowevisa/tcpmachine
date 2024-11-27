package tcpserver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"git.qowevisa.me/qowevisa/tcpmachine/tcpcommand"
)

const (
	LogLevel_Nothing    = 0
	LogLevel_Connection = 1 << iota
	LogLevel_Messages
)

const (
	LogLevel_ALL = LogLevel_Connection | LogLevel_Messages
)

func CreateHandleClientFuncFromCommands(bundle *tcpcommand.CommandBundle, conf ServerConfiguration) (func(client net.Conn), chan error) {
	clientErrors := make(chan error, 16)
	return func(client net.Conn) {
		defer client.Close()
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

func defaultHandleClientFunc(server *Server) func(net.Conn) {
	return func(client net.Conn) {
		if server.PostHandlerClientFunc != nil {
			defer server.PostHandlerClientFunc(client)
		}
		defer client.Close()
		connReader := bufio.NewReader(client)
		for {
			msg, err := connReader.ReadString(byte(server.MessageEndRune))
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				server.ErrorsChannel <- err
			}
			msgWoNl := strings.Trim(msg, string(server.MessageEndRune))
			parts := strings.Split(msgWoNl, string(server.MessageSplitRune))
			commandFound := false
			if server.LogLevel&LogLevel_Messages > 0 {
				log.Printf("Message received from %s : %s\n", client.RemoteAddr(), msgWoNl)
			}
			for _, cmd := range server.Commands {
				if cmd.Command == parts[0] {
					cmd.Action(parts[1:], client)
					commandFound = true
					break
				}
			}
			if !commandFound {
				server.ErrorsChannel <- fmt.Errorf("WARN: Command %s error: %w", parts[0], commandNotHandledError)
			}
		}
	}
}

// NOTE: HandleClientFunc is NOT created by this function
// see: CreateHandleClientFuncFromCommands(bundle)
func GetDefaultConfig() ServerConfiguration {
	return ServerConfiguration{
		MessageEndRune:   '\n',
		MessageSplitRune: ' ',
		ErrorResolver: func(c chan error) {
			for err := range c {
				log.Printf("DefConfig:Error: %v\n", err)
			}
		},
	}
}

func CreateServer(addr string, options ...ServerOption) *Server {
	conf := GetDefaultConfig()

	for _, opt := range options {
		opt(&conf)
	}
	var cmds []tcpcommand.Command

	server := &Server{
		addr:             addr,
		Exit:             make(chan bool, 1),
		HandleClientFunc: conf.HandleClientFunc,
		//
		MessageEndRune:   conf.MessageEndRune,
		MessageSplitRune: conf.MessageSplitRune,
		ErrorsChannel:    make(chan error, 8),
		ErrorResolver:    conf.ErrorResolver,
		LogLevel:         conf.LogLevel,
		//
		Commands: cmds,
	}
	if conf.HandleClientFunc == nil {
		server.HandleClientFunc = defaultHandleClientFunc(server)
	}

	return server
}

func (s *Server) StartServer() error {
	if s.Exit == nil {
		return fmt.Errorf("server's Exit channel is nil. Can't start server")
	}
	if s.HandleClientFunc == nil {
		return fmt.Errorf("server's HandleClientFunc is nil. Can't start server")
	}
	if s.ErrorsChannel == nil {
		return fmt.Errorf("server's ErrorsChannel is nil. Can't start server")
	}
	if len(s.Commands) == 0 {
		s.ErrorsChannel <- fmt.Errorf("WARN: len(s.Commands) == 0! Server is running without commands")
	}
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}
	defer listener.Close()
	if s.ErrorResolver == nil {
		return fmt.Errorf("server's ErrorResolver is nil. Can't start server")
	}
	go s.ErrorResolver(s.ErrorsChannel)
	log.Printf("Server started listening on %s\n", s.addr)

loop:
	for {
		select {
		case <-s.Exit:
			break loop
		default:
			newClient, err := listener.Accept()
			if err != nil {
				s.ErrorsChannel <- fmt.Errorf("listener.Accept: %w", err)
				break
			}
			if s.LogLevel&LogLevel_Connection > 0 {
				fmt.Printf("New Connection from %s is accepted\n", newClient.RemoteAddr())
			}
			if s.PreHandlerClientFunc != nil {
				s.PreHandlerClientFunc(newClient)
			}
			go s.HandleClientFunc(newClient)
		}
	}
	return nil
}

var commandDuplicateError = errors.New("Command already exists in server")
var commandNotHandledError = errors.New("Command was not handled")

// On registers an action for a specific command. The action will be executed
// when the command is received from a client. For example, given the input
// string `TEST 1 2 3 4`, the command would be `TEST`, and the args would be
// []string{"1", "2", "3", "4"}.
//
// NOTE: This behavior applies only if `MessageSplitRune` is set to the default value (' ').
func (s *Server) On(command string, action func(args []string, client net.Conn)) error {
	for _, cmd := range s.Commands {
		if cmd.Command == command {
			return fmt.Errorf("Failed addding command %s: %w ", command, commandDuplicateError)
		}
	}
	s.Commands = append(s.Commands, tcpcommand.Command{
		Command: command,
		Action:  action,
	})
	return nil
}
