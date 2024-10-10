package tcpcommand

import (
	"errors"
	"net"
)

type Command struct {
	// Command is what server/client will RECEIVE
	Command string
	// Action is what to DO when such command is RECEIVED
	Action func([]string, net.Conn)
}

type CommandBundle struct {
	Commands []Command
}

// implementation is O(n)
func (cb *CommandBundle) alreadyHaveCommand(cmd string) bool {
	for _, command := range cb.Commands {
		if command.Command == cmd {
			return true
		}
	}
	return false
}

var ErrCommandBundleDuplicateCommands = errors.New("Found two commands with the same Command value")

func CreateCommandBundle(commands []Command) (*CommandBundle, error) {
	cb := &CommandBundle{
		Commands: []Command{},
	}
	for _, cmd := range commands {
		// cb.alreadyHaveCommand is O(n) but it's ok since we won't have commandBundle with over 100k commands
		if cb.alreadyHaveCommand(cmd.Command) {
			return nil, ErrCommandBundleDuplicateCommands
		}
		cb.Commands = append(cb.Commands, cmd)
	}
	return cb, nil
}
