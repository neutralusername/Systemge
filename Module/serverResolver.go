package Module

import (
	"Systemge/TopicResolutionServer"
	"Systemge/Utilities"
	"strings"
)

func NewResolverServer(name string, port string, loggerPath string, topicAddresses map[string]string) *TopicResolutionServer.Server {
	logger := Utilities.NewLogger(loggerPath)
	resolutionServer := TopicResolutionServer.New(name, port, logger)
	for topic, address := range topicAddresses {
		resolutionServer.RegisterTopics(address, topic)
	}
	return resolutionServer
}

func NewResolverServerFromConfig(sytemgeConfigPath string, errorLogPath string) *TopicResolutionServer.Server {
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
	topics := map[string]string{}
	lines := Utilities.SplitLines(Utilities.GetFileContent(sytemgeConfigPath))
	if len(lines) < 2 {
		panic("error reading file")
	}
	for i, line := range lines {
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
			if segments[1] != "resolver" {
				panic("error reading file. invalid config type (want: resolver)")
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
			segments := strings.Split(line, " ")
			if len(segments) != 2 {
				panic("error reading file. incomplete topic/address pair")
			}
			if len(segments[0]) < 1 || len(segments[1]) < 1 {
				panic("error reading file. Missing topic or address")
			}
			if segments[1][0] != ':' {
				panic("error reading file. Missing port number")
			}
			topics[segments[0]] = segments[1]
		}
	}
	return NewResolverServer(name, port, errorLogPath, topics)
}
