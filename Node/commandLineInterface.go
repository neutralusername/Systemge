package Node

import (
	"Systemge/Helpers"
	"Systemge/Tools"
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
	println("enter command (exit to quit) (? for help)")
	for {
		print(">")
		input, err := reader.ReadString(byte(newLineChar))
		if err != nil {
			println(err.Error())
			continue
		}
		input = strings.Trim(input, "\r\n")
		inputSegments := strings.Split(input, " ")
		if inputSegments[0] == "" || inputSegments[0] == "!" {
			continue
		}
		reverse := false
		if inputSegments[0][0] == '!' {
			inputSegments[0] = inputSegments[0][1:]
			reverse = true
		}
		switch inputSegments[0] {
		case "?":
			println("start the command with ! to reverse the order of the nodes in which they were provided")
			println("startSchedule <command> <time in ms> <repeat> <args...>")
			println("stopSchedule <scheduleId>")
			println("listSchedules")
			println("restart stops all nodes and then starts them again in the order they were provided (! affects only the stop part)")
		case "exit":
			return
		case "startSchedule":
			startSchedule(inputSegments, schedules, nodes...)
		case "stopSchedule":
			stopSchedule(inputSegments, schedules)
		case "listSchedules":
			for scheduleId, schedule := range schedules {
				fmt.Println("scheduleId: \"" + scheduleId + "\", command: \"" + schedule.command + "\" time started: " + schedule.timeStarted.String() + ", duration: " + schedule.duration.String() + ", repeat: " + fmt.Sprint(schedule.repeat) + ", args: " + strings.Join(schedule.args, " "))
			}
		case "restart":
			restart(reverse, nodes...)
		default:
			handleCommands(reverse, inputSegments, nodes...)
		}
	}
}

func restart(reverse bool, nodes ...*Node) {
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
	scheduleId := Tools.RandomString(10, Tools.ALPHA_NUMERIC)
	for schedules[scheduleId] != nil {
		scheduleId = Tools.RandomString(10, Tools.ALPHA_NUMERIC)
	}
	schedules[scheduleId] = schedule
	println("schedule started with id \"" + scheduleId + "\"")
	schedule.timer = time.AfterFunc(schedule.duration, func() {
		reverse := false
		c := command
		if command[0] == '!' {
			reverse = true
			c = command[1:]
		}
		if c == "restart" {
			restart(reverse, nodes...)
		} else {
			handleCommands(reverse, append([]string{c}, args...), nodes...)
		}
		if schedule.repeat {
			schedule.timer.Reset(schedule.duration)
		} else {
			delete(schedules, scheduleId)
		}
	})
}

func handleCommands(reverse bool, inputSegments []string, nodes ...*Node) bool {
	commandExecuted := false
	if reverse {
		for i := len(nodes) - 1; i >= 0; i-- {
			executeCommand(nodes[i], inputSegments[0], &commandExecuted, inputSegments[1:]...)
		}
	} else {
		for _, node := range nodes {
			executeCommand(node, inputSegments[0], &commandExecuted, inputSegments[1:]...)
		}
	}
	if !commandExecuted {
		println("command not found \"" + inputSegments[0] + "\"")
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
