package Module

import (
	"Systemge/Resolver"
	"Systemge/Utilities"
	"strings"
)

func NewResolver(name, resolverPort, configPort, tlsCertPath, tlsKeyPath, loggerPath string, brokers map[string]*Resolver.Resolution, topics map[string]*Resolver.Resolution) *Resolver.Server {
	resolver := Resolver.New(name, resolverPort, configPort, tlsCertPath, tlsKeyPath, Utilities.NewLogger(loggerPath))
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
	resolverPort := ""
	configPort := ""
	tlsCertPath := ""
	tlsKeyPath := ""
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
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. incomplete port number")
			}
			if lineSegments[0] != "resolver" {
				panic("error reading file. invalid port number")
			}
			if len(lineSegments[1]) < 2 {
				panic("error reading file. Missing port number")
			}
			if lineSegments[1][0] != ':' {
				panic("error reading file. Missing port number")
			}
			resolverPort = lineSegments[1]
		case 2:
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. incomplete port number")
			}
			if lineSegments[0] != "config" {
				panic("error reading file. invalid port number")
			}
			if len(lineSegments[1]) < 2 {
				panic("error reading file. Missing port number")
			}
			if lineSegments[1][0] != ':' {
				panic("error reading file. Missing port number")
			}
			configPort = lineSegments[1]
		case 3:
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. invalid cert or key line")
			}
			if lineSegments[0] == "cert" {
				tlsCertPath = lineSegments[1]
			} else if lineSegments[0] == "key" {
				tlsKeyPath = lineSegments[1]
			} else {
				panic("error reading file. invalid cert or key type")
			}
		case 4:
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. invalid cert or key line")
			}
			if lineSegments[0] == "cert" {
				tlsCertPath = lineSegments[1]
			} else if lineSegments[0] == "key" {
				tlsKeyPath = lineSegments[1]
			} else {
				panic("error reading file. invalid cert or key type")
			}
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
				if !Utilities.FileExists(lineSegments[2]) {
					panic("error reading file. invalid cert path")
				}
				cert := Utilities.GetFileContent(lineSegments[2])
				broker := Resolver.NewResolution(lineSegments[0], lineSegments[1], cert)
				brokers[lineSegments[0]] = broker
			} else {
				panic("error reading file. invalid topic line")
			}

		}
	}
	return NewResolver(name, resolverPort, configPort, tlsCertPath, tlsKeyPath, errorLogPath, brokers, topics)
}
