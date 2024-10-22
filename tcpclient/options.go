package tcpclient

import "net"

type ClientOption func(conf *ClientConfiguration)

const (
	statusBitNothing             = 0
	statusBitCustomErrorHandling = 1 << iota
)

func WithCustomErrorHandling(fun ErrorResolverFunc) ClientOption {
	return func(conf *ClientConfiguration) {
		conf.Status |= statusBitCustomErrorHandling
		conf.ErrorResolver = fun
	}
}

func WithServerHandler(fun func(server net.Conn)) ClientOption {
	return func(conf *ClientConfiguration) {
		conf.ServerHandlerFunc = fun
	}
}
