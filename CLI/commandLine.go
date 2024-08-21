package CLI

import (
	"fmt"
	"strings"

	"github.com/neutralusername/Systemge/Dashboard"
)

func Start(commandHandlers Dashboard.CommandHandlers) {
	println("Welcome to the Systemge command line interface")
	for {
		print("> ")
		var command string
		_, _ = fmt.Scanln(&command)
		if command == "exit" {
			break
		}
		segments := strings.Split(command, " ")
		if len(segments) == 0 {
			continue
		}
		commandHandler, ok := commandHandlers[segments[0]]
		if !ok {
			println("Command not found")
			continue
		}
		commandHandler(segments[1:])
	}
}
