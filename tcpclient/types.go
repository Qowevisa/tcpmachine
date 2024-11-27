package tcpclient

import (
	"net"

	"git.qowevisa.me/qowevisa/tcpmachine/tcpcommand"
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
