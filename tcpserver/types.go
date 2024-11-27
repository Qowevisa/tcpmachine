package tcpserver

import (
	"net"

	"git.qowevisa.me/qowevisa/tcpmachine/tcpcommand"
)

type ServerLoggingLevel int

type ServerConfiguration struct {
	MessageEndRune   rune
	MessageSplitRune rune
	HandleClientFunc func(client net.Conn)
	LogLevel         ServerLoggingLevel
	//
	ErrorResolver func(chan error)
}

type Server struct {
	addr                 string
	PreHandlerClientFunc func(client net.Conn)
	HandleClientFunc     func(client net.Conn)
	// Use PostHandlerClientFunc in your HandleClientFunc if you
	// use custom HandleClientFunc
	PostHandlerClientFunc func(client net.Conn)
	Exit                  chan bool
	//
	MessageEndRune   rune
	MessageSplitRune rune
	ErrorsChannel    chan error
	ErrorResolver    func(chan error)
	LogLevel         ServerLoggingLevel
	//
	Commands []tcpcommand.Command
}
