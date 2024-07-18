package Node

import (
	"bufio"
	"os"
	"runtime"
	"strings"
)

// starts a command-line interface for the module
func StartCommandLineInterface(stopReversedOrder bool, nodes ...*Node) {
	if len(nodes) == 0 {
		panic("no nodes provided")
	}
	newLineChar := '\n'
	if runtime.GOOS == "windows" {
		newLineChar = '\r'
	}
	reader := bufio.NewReader(os.Stdin)
	started := false
	println("enter command (exit to quit)")
	for {
		print(">")
		input, err := reader.ReadString(byte(newLineChar))
		if err != nil {
			continue
		}
		input = strings.Trim(input, "\r\n")
		inputSegments := strings.Split(input, " ")
		switch inputSegments[0] {

		case "start":
			if started {
				println("module already started")
				continue
			}
			for _, node := range nodes {
				err := node.Start()
				if err != nil {
					panic(err.Error())
				}
			}
			started = true
		case "stop":
			if !started {
				println("module not started")
				continue
			}
			for _, node := range nodes {
				err := node.Stop()
				if err != nil {
					panic(err.Error())
				}
			}
			started = false
		case "exit":
			return
		default:
			if !started {
				println("module not started")
				continue
			}
			commandExecuted := false
			for _, node := range nodes {
				commandHandlers := node.GetCommandHandlers()
				handler := commandHandlers[inputSegments[0]]
				if handler != nil {
					commandExecuted = true
					err := handler(node, inputSegments[1:])
					if err != nil {
						println(err.Error())
					}
				}
			}
			if !commandExecuted {
				println("command not found")
			}
		}
	}
}
