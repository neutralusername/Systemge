package Node

import (
	"bufio"
	"os"
	"runtime"
	"strings"
)

// starts a command-line interface for the provided nodes
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
				println("cli already started")
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
				println("cli not started")
				continue
			}
			if stopReversedOrder {
				for i := len(nodes) - 1; i >= 0; i-- {
					err := nodes[i].Stop()
					if err != nil {
						panic(err.Error())
					}
				}
			} else {
				for _, node := range nodes {
					err := node.Stop()
					if err != nil {
						panic(err.Error())
					}
				}
			}
			started = false
		case "exit":
			return
		default:
			if !started {
				println("cli not started")
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
