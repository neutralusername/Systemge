package Module

import (
	"Systemge/MessageBrokerServer"
	"Systemge/Utilities"
	"strings"
)

func NewBrokerServer(name string, port string, loggerPath string, topics ...string) *MessageBrokerServer.Server {
	logger := Utilities.NewLogger(loggerPath)
	messageBrokerServer := MessageBrokerServer.New(name, port, logger)
	for _, topic := range topics {
		messageBrokerServer.AddTopics(topic)
	}
	return messageBrokerServer
}

func NewBrokerServerFromConfig(sytemgeConfigPath string, errorLogPath string) *MessageBrokerServer.Server {
	if !Utilities.FileExists(sytemgeConfigPath) {
		panic("provided file does not exist")
	}
	filename := Utilities.GetFileName(sytemgeConfigPath)
	fileNameSegments := strings.Split(Utilities.GetFileName(sytemgeConfigPath), ".")
	name := ""
	if len(fileNameSegments) != 2 {
		name = filename
	} else {
		name = fileNameSegments[0]
	}
	port := ""
	topics := []string{}
	for i, line := range Utilities.SplitLines(Utilities.GetFileContent(sytemgeConfigPath)) {
		if len(line) == 0 {
			continue
		}
		switch i {
		case 0:
			if len(line) < 2 {
				panic("error reading file. missing config type")
			}
			segments := strings.Split(line, " ")
			if len(segments) != 2 {
				panic("error reading file. incomplete config type")
			}
			if segments[0] != "#" {
				panic("error reading file. missing config type identifier (#)")
			}
			if segments[1] != "broker" {
				panic("error reading file. invalid config type (want: broker)")
			}
			continue
		case 1:
			if len(line) < 2 {
				panic("error reading file. Missing port number")
			}
			if line[0] != ':' {
				panic("error reading file. Missing port number")
			}
			port = line
		default:
			topics = append(topics, line)
		}
	}
	return NewBrokerServer(name, port, errorLogPath, topics...)
}
