package Node

import (
	"Systemge/Helpers"
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type Schedule struct {
	timer       *time.Timer
	timeStarted time.Time
	duration    time.Duration
	repeat      bool
	args        []string
	command     string
}

// starts a command-line interface for the provided nodes
func StartCommandLineInterface(nodes ...*Node) {
	if len(nodes) == 0 {
		panic("no nodes provided")
	}
	newLineChar := '\n'
	if runtime.GOOS == "windows" {
		newLineChar = '\r'
	}
	reader := bufio.NewReader(os.Stdin)
	schedules := map[string]*Schedule{}
	println("enter command (exit to quit)")
	for {
		print(">")
		input, err := reader.ReadString(byte(newLineChar))
		if err != nil {
			continue
		}
		input = strings.Trim(input, "\r\n")
		inputSegments := strings.Split(input, " ")
		if inputSegments[0] == "" {
			continue
		}
		reverse := false
		if inputSegments[0][0] == '!' {
			inputSegments[0] = inputSegments[0][1:]
			reverse = true
		}
		switch inputSegments[0] {
		case "exit":
			return
		case "startSchedule":
			startSchedule(inputSegments, schedules, nodes...)
		case "stopSchedule":
			stopSchedule(inputSegments, schedules)
		case "listSchedules":
			for command, schedule := range schedules {
				fmt.Println("command: \"" + command + "\", time started: " + schedule.timeStarted.String() + ", duration: " + schedule.duration.String() + ", repeat: " + fmt.Sprint(schedule.repeat) + ", args: " + strings.Join(schedule.args, " "))
			}
		case "restart":
			if reverse {
				for i := len(nodes) - 1; i >= 0; i-- {
					executeCommand(nodes[i], "stop", nil)
				}
			} else {
				for _, node := range nodes {
					executeCommand(node, "stop", nil)
				}
			}
			for _, node := range nodes {
				executeCommand(node, "start", nil)
			}
		default:
			handleCommands(reverse, inputSegments, nodes...)
		}
	}
}

func stopSchedule(inputSegments []string, schedules map[string]*Schedule) {
	if len(inputSegments) < 2 {
		println("invalid number of arguments")
		return
	}
	command := inputSegments[1]
	schedule := schedules[command]
	if schedule == nil {
		println("schedule not found")
		return
	}
	schedule.timer.Stop()
	delete(schedules, command)
}

func startSchedule(inputSegments []string, schedules map[string]*Schedule, nodes ...*Node) {
	if len(inputSegments) < 4 {
		println("invalid number of arguments")
		return
	}
	if schedules[inputSegments[1]] != nil {
		stopSchedule([]string{"stopSchedule", inputSegments[1]}, schedules)
	}
	command := inputSegments[1]
	timeMs := Helpers.StringToUint64(inputSegments[2])
	repeat := inputSegments[3] == "true"
	args := inputSegments[4:]
	schedule := &Schedule{
		timeStarted: time.Now(),
		duration:    time.Duration(timeMs) * time.Millisecond,
		repeat:      repeat,
		args:        args,
		command:     command,
	}
	schedules[command] = schedule
	schedule.timer = time.AfterFunc(schedule.duration, func() {
		reverse := false
		c := command
		if command[0] == '!' {
			reverse = true
			c = command[1:]
		}
		handleCommands(reverse, append([]string{c}, args...), nodes...)
		if schedule.repeat {
			schedule.timer.Reset(schedule.duration)
		} else {
			delete(schedules, command)
		}
	})
}

func handleCommands(reverse bool, inputSegments []string, nodes ...*Node) bool {
	commandExecuted := false
	if reverse {
		for i := len(nodes) - 1; i >= 0; i-- {
			executeCommand(nodes[i], inputSegments[0], &commandExecuted, inputSegments[1:]...)
		}
		return true
	} else {
		for _, node := range nodes {
			executeCommand(node, inputSegments[0], &commandExecuted, inputSegments[1:]...)
		}
	}
	if !commandExecuted {
		println("command not found")
	}
	return true
}

func executeCommand(node *Node, command string, commandExecuted *bool, args ...string) {
	commandHandlers := node.GetCommandHandlers()
	handler := commandHandlers[command]
	if handler != nil {
		if commandExecuted != nil {
			*commandExecuted = true
		}
		println("\texecuting command \"" + command + "\" on node \"" + node.GetName() + "\"")
		err := handler(node, args)
		if err != nil {
			println(err.Error())
		}
	}
}
