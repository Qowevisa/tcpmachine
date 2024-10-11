package tcpserver

import "net"

type ServerOption func(*ServerConfiguration)

// WithMessageEndRune sets the MessageEndRune in the server configuration.
func WithMessageEndRune(r rune) ServerOption {
	return func(conf *ServerConfiguration) {
		conf.MessageEndRune = r
	}
}

// WithMessageSplitRune sets the MessageSplitRune in the server configuration.
func WithMessageSplitRune(r rune) ServerOption {
	return func(conf *ServerConfiguration) {
		conf.MessageSplitRune = r
	}
}

// WithErrorResolver sets a custom error resolver function.
func WithErrorResolver(resolver func(chan error)) ServerOption {
	return func(conf *ServerConfiguration) {
		conf.ErrorResolver = resolver
	}
}

// WithHandleClientFunc sets the HandleClientFunc in the server configuration.
func WithHandleClientFunc(handler func(client net.Conn)) ServerOption {
	return func(conf *ServerConfiguration) {
		conf.HandleClientFunc = handler
	}
}
