package CLI

import (
	"fmt"
	"strings"

	"github.com/neutralusername/Systemge/Commands"
)

func Start(commandHandlers Commands.Handlers) {
	println("Welcome to the Systemge command line interface")
	println("Type 'exit' to exit")
	println("Type your command followed by arguments separated by spaces")
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
		result, err := commandHandler(segments[1:])
		if err != nil {
			println("Error: " + err.Error())
			continue
		}
		println(result)
	}
}
