ping-pong: pp-server pp-client
	@

pp-client pp-server:
	go build -o ./bin/$@ ./examples/ping-pong/$@/
