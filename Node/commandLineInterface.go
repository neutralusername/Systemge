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
	isStarted := false

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
		case "restart":
			if stop(&isStarted, stopReversedOrder, nodes...) {
				start(&isStarted, nodes...)
			}
		case "start":
			start(&isStarted, nodes...)
		case "stop":
			stop(&isStarted, stopReversedOrder, nodes...)
		case "exit":
			return
		default:
			handleCommands(isStarted, inputSegments, nodes...)
		}
	}
}

func handleCommands(isStarted bool, inputSegments []string, nodes ...*Node) bool {
	if !isStarted {
		println("cli not started")
		return false
	}
	commandExecuted := false
	for _, node := range nodes {
		commandHandlers := node.GetCommandHandlers()
		handler := commandHandlers[inputSegments[0]]
		if handler != nil {
			commandExecuted = true
			println("\texecuting command on node \"" + node.GetName() + "\"")
			err := handler(node, inputSegments[1:])
			if err != nil {
				println(err.Error())
			}
		}
	}
	if !commandExecuted {
		println("command not found")
	}
	return true
}

func start(isStarted *bool, nodes ...*Node) bool {
	if *isStarted {
		println("cli already started")
		return false
	}
	for _, node := range nodes {
		err := node.Start()
		if err != nil {
			panic(err.Error())
		}
	}
	*isStarted = true
	return true
}

func stop(isStarted *bool, stopReversedOrder bool, nodes ...*Node) bool {
	if !*isStarted {
		println("cli not started")
		return false
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
	*isStarted = false
	return true
}
