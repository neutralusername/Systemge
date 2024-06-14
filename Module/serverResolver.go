package Module

import (
	"Systemge/ResolverServer"
	"Systemge/Utilities"
	"strings"
)

func NewResolverServer(name string, port string, loggerPath string, brokers map[string]*ResolverServer.Resolution, topics map[string]*ResolverServer.Resolution) *ResolverServer.Server {
	logger := Utilities.NewLogger(loggerPath)
	resolutionServer := ResolverServer.New(name, port, logger)
	for _, broker := range brokers {
		resolutionServer.RegisterBroker(broker)
	}
	for topic, broker := range topics {
		resolutionServer.RegisterTopics(broker.Name, topic)
	}
	return resolutionServer
}

func NewResolverServerFromConfig(sytemgeConfigPath string, errorLogPath string) *ResolverServer.Server {
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
	topics := map[string]*ResolverServer.Resolution{}  // topic -> broker
	brokers := map[string]*ResolverServer.Resolution{} // broker-name -> broker
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
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. incomplete config type")
			}
			if lineSegments[0] != "#" {
				panic("error reading file. missing config type identifier (#)")
			}
			if lineSegments[1] != "resolver" {
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
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) == 2 {
				if len(lineSegments[0]) < 1 || len(lineSegments[1]) < 1 {
					panic("error reading file. Missing topic or address")
				}
				broker := brokers[lineSegments[1]]
				if broker == nil {
					panic("error reading file. Broker not found")
				}
				topics[lineSegments[0]] = broker
			} else if len(lineSegments) == 3 {
				name := lineSegments[0]
				address := lineSegments[1]
				if !Utilities.FileExists(lineSegments[2]) {
					println(lineSegments[2])
					panic("error reading file. cert file not found")
				}
				cert := Utilities.GetFileContent(lineSegments[2])
				broker := ResolverServer.NewResolution(name, address, cert)
				brokers[name] = broker
			} else {
				panic("error reading file. invalid topic line")
			}

		}
	}
	return NewResolverServer(name, port, errorLogPath, brokers, topics)
}
