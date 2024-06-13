package Module

import (
	"Systemge/Application"
	"bufio"
	"os"
	"runtime"
	"strings"
)

// starts a command-line interface for the module
func StartCommandLineInterface(module Module, customCommands map[string]Application.CustomCommandHandler) {
	if module == nil {
		panic("module cannot be nil")
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
			err := module.Start()
			if err != nil {
				panic(err.Error())
			}
			started = true
		case "stop":
			if !started {
				println("module not started")
				continue
			}
			err := module.Stop()
			if err != nil {
				panic(err.Error())
			}
			started = false
		case "exit":
			return
		default:
			if customCommands != nil && customCommands[inputSegments[0]] != nil {
				if !started {
					println("module not started")
					continue
				}
				err := customCommands[inputSegments[0]](inputSegments[1:])
				if err != nil {
					println("error executing command: " + err.Error())
				}
				continue
			} else {
				println("unknown command \"" + input + "\"")
			}
		}
	}
}
