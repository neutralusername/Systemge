package Module

import (
	"bufio"
	"os"
	"runtime"
	"strings"
)

// starts a command-line interface for the module
func StartCommandLineInterface(module Module) {
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
			if !started {
				println("module not started")
				continue
			}
			customCommandHandlers := module.GetCommandHandlers()
			handler := customCommandHandlers[inputSegments[0]]
			if handler != nil {
				err := handler(inputSegments[1:])
				if err != nil {
					println(err.Error())
				}
			} else {
				println("command not found")
			}
		}
	}
}
