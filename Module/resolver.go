package Module

import (
	"Systemge/Resolver"
	"Systemge/Utilities"
	"strings"
)

func NewResolver(name string, port string, loggerPath string, brokers map[string]*Resolver.Resolution, topics map[string]*Resolver.Resolution) *Resolver.Server {
	logger := Utilities.NewLogger(loggerPath)
	resolver := Resolver.New(name, port, logger)
	for _, broker := range brokers {
		resolver.RegisterBroker(broker)
	}
	for topic, broker := range topics {
		resolver.RegisterTopics(broker.Name, topic)
	}
	return resolver
}

func NewResolverFromConfig(sytemgeConfigPath string, errorLogPath string) *Resolver.Server {
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
	topics := map[string]*Resolver.Resolution{}  // topic -> broker
	brokers := map[string]*Resolver.Resolution{} // broker-name -> broker
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
				broker := Resolver.NewResolution(name, address, cert)
				brokers[name] = broker
			} else {
				panic("error reading file. invalid topic line")
			}

		}
	}
	return NewResolver(name, port, errorLogPath, brokers, topics)
}
