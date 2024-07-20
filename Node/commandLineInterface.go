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
	nodeNames   []string
}

func processSegments(inputSegments []string) (nodeNames []string, reverse bool, command string, args []string) {
	nodeNames = []string{}
	for _, segment := range inputSegments {
		if segment[0] != '@' {
			break
		}
		nodeNames = append(nodeNames, segment[1:])
	}
	inputSegments = inputSegments[len(nodeNames):]
	reverse = false
	if inputSegments[0][0] == '!' {
		inputSegments[0] = inputSegments[0][1:]
		reverse = true
	}
	command = inputSegments[0]
	args = inputSegments[1:]
	return
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
		if inputSegments[0] == "" || inputSegments[0] == "!" || inputSegments[0] == "@" {
			continue
		}
		nodeNames, reverse, command, args := processSegments(inputSegments)
		switch command {
		case "?":
			println("> commands from the custom command handlers of nodes can be executed in this command line interface")
			println("> type <listNodes> to see the available nodes")
			println("> type <listCommands> to see the available commands for each node")
			println("> precede any command with a variable amount of <@nodeName> to execute the command on specific nodes")
			println("> precede any command with an exclamation mark (<!command>) to reverse the order of nodes for the command")
			println("> schedule a command to be executed at a later time")
			println("> after starting a schedule its scheduleId will be returned which can be used to stop the schedule")
			println("> startSchedule <command> <time in ms> <repeat> <args...>")
			println("> restart stops all nodes and then starts them again in the order they were provided (! affects only the stop part)")
		case "exit":
			return
		case "listCommands":
			for _, node := range nodes {
				println("node: \"" + node.GetName() + "\"")
				for command := range node.GetCommandHandlers() {
					println("\t\"" + command + "\"")
				}
			}
			println("schedule commands: startSchedule, stopSchedule, listSchedules")
			println("system commands: listNodes, restart")
		case "listNodes":
			for _, node := range nodes {
				println("node: \"" + node.GetName() + "\"")
			}
		case "startSchedule":
			startSchedule(inputSegments, schedules, nodes...)
		case "stopSchedule":
			stopSchedule(inputSegments, schedules)
		case "listSchedules":
			for scheduleId, schedule := range schedules {
				fmt.Println("scheduleId: \"" + scheduleId + "\", command: \"" + schedule.command + "\" time started: " + schedule.timeStarted.String() + ", duration: " + schedule.duration.String() + ", repeat: " + fmt.Sprint(schedule.repeat) + ", nodeNames: " + fmt.Sprint(schedule.nodeNames) + ", args: " + fmt.Sprint(schedule.args))
			}
		case "restart":
			restart(nodeNames, reverse, nodes...)
		default:
			handleCommands(nodeNames, reverse, command, args, nodes...)
		}
	}
}

func startSchedule(inputSegments []string, schedules map[string]*Schedule, nodes ...*Node) {
	if len(inputSegments) < 4 {
		println("invalid number of arguments")
		return
	}
	if schedules[inputSegments[1]] != nil {
		stopSchedule([]string{"stopSchedule", inputSegments[1]}, schedules)
	}
	nodeNames, reverse, _, args := processSegments(inputSegments)
	command := args[0]
	timeMs := Helpers.StringToInt(args[1])
	repeat := args[2] == "true"
	args = args[3:]
	fmt.Println(nodeNames, reverse, command, args)
	schedule := &Schedule{
		timeStarted: time.Now(),
		duration:    time.Duration(timeMs) * time.Millisecond,
		repeat:      repeat,
		args:        args,
		command:     command,
		nodeNames:   nodeNames,
	}
	scheduleId := Tools.RandomString(10, Tools.ALPHA_NUMERIC)
	for schedules[scheduleId] != nil {
		scheduleId = Tools.RandomString(10, Tools.ALPHA_NUMERIC)
	}
	schedules[scheduleId] = schedule
	println("schedule started with id \"" + scheduleId + "\"")
	schedule.timer = time.AfterFunc(schedule.duration, func() {
		if command == "restart" {
			restart(nodeNames, reverse, nodes...)
		} else {
			handleCommands(nodeNames, reverse, command, args, nodes...)
		}
		if schedule.repeat {
			schedule.timer.Reset(schedule.duration)
		} else {
			delete(schedules, scheduleId)
		}
	})
}

func restart(nodeNames []string, reverse bool, nodes ...*Node) {
	if reverse {
		if len(nodeNames) > 0 {
			for i := len(nodeNames) - 1; i >= 0; i-- {
				for _, node := range nodes {
					if node.GetName() == nodeNames[i] {
						executeCommand(node, "stop", nil)
					}
				}
			}
		} else {
			for i := len(nodes) - 1; i >= 0; i-- {
				executeCommand(nodes[i], "stop", nil)
			}
		}
	} else {
		if len(nodeNames) > 0 {
			for _, nodeName := range nodeNames {
				for _, node := range nodes {
					if node.GetName() == nodeName {
						executeCommand(node, "stop", nil)
					}
				}
			}
		} else {
			for _, node := range nodes {
				executeCommand(node, "stop", nil)
			}
		}
	}
	if len(nodeNames) > 0 {
		for _, nodeName := range nodeNames {
			for _, node := range nodes {
				if node.GetName() == nodeName {
					executeCommand(node, "start", nil)
				}
			}
		}
	} else {
		for _, node := range nodes {
			executeCommand(node, "start", nil)
		}
	}
}

func handleCommands(nodeNames []string, reverse bool, command string, args []string, nodes ...*Node) bool {
	commandExecuted := false
	if reverse {
		if len(nodeNames) > 0 {
			for i := len(nodeNames) - 1; i >= 0; i-- {
				for _, node := range nodes {
					if node.GetName() == nodeNames[i] {
						executeCommand(node, command, &commandExecuted, args...)
					}
				}
			}
		} else {
			for i := len(nodes) - 1; i >= 0; i-- {
				executeCommand(nodes[i], command, &commandExecuted, args...)
			}
		}
	} else {
		if len(nodeNames) > 0 {
			for _, nodeName := range nodeNames {
				for _, node := range nodes {
					if node.GetName() == nodeName {
						executeCommand(node, command, &commandExecuted, args...)
					}
				}
			}
		} else {
			for _, node := range nodes {
				executeCommand(node, command, &commandExecuted, args...)
			}
		}
	}
	if !commandExecuted {
		println("command not found \"" + command + "\"")
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
