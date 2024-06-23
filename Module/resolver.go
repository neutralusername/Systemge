package Module

import (
	"Systemge/Resolution"
	"Systemge/Resolver"
	"Systemge/Utilities"
	"strings"
)

func NewResolver(name, resolverPort, resolverTlsCertPath, resolverTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath, loggerPath string, brokers map[string]*Resolution.Resolution, topics map[string]*Resolution.Resolution) *Resolver.Server {
	resolver := Resolver.New(name, resolverPort, resolverTlsCertPath, resolverTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath, Utilities.NewLogger(loggerPath))
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
		panic("provided file does not exist \"" + sytemgeConfigPath + "\"")
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
	resolverTlsCertPath := ""
	resolverTlsKeyPath := ""
	configPort := ""
	configTlsCertPath := ""
	configTlsKeyPath := ""
	topics := map[string]*Resolution.Resolution{}  // topic -> broker
	brokers := map[string]*Resolution.Resolution{} // broker-name -> broker
	lines := Utilities.SplitLines(Utilities.GetFileContent(sytemgeConfigPath))
	if len(lines) < 3 {
		panic("provided file has too few lines to be a valid config")
	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		lineSegmentss := strings.Split(line, " ")
		lineSegments := []string{}
		for _, segment := range lineSegmentss {
			if len(segment) > 0 {
				lineSegments = append(lineSegments, segment)
			}
		}
		switch lineSegments[0] {
		case "#":
			if lineSegments[1] != "resolver" {
				panic("wrong config type for resolver \"" + lineSegments[1] + "\"")
			}
		case "resolver":
			if len(lineSegments) != 4 {
				panic("resolver line is invalid \"" + line + "\"")
			}
			resolverPort = lineSegments[1]
			resolverTlsCertPath = lineSegments[2]
			resolverTlsKeyPath = lineSegments[3]
		case "config":
			if len(lineSegments) != 4 {
				panic("config line is invalid \"" + line + "\"")
			}
			configPort = lineSegments[1]
			configTlsCertPath = lineSegments[2]
			configTlsKeyPath = lineSegments[3]
		default:
			if len(lineSegments) < 4 {
				panic("broker line is invalid \"" + line + "\"")
			}
			if !Utilities.FileExists(lineSegments[3]) {
				panic("certificate file does not exist \"" + lineSegments[2] + "\"")
			}
			resolution := Resolution.New(lineSegments[0], lineSegments[1], lineSegments[2], Utilities.GetFileContent(lineSegments[3]))
			brokers[lineSegments[0]] = resolution
			for _, topic := range lineSegments[3:] {
				topics[topic] = resolution
			}
		}
	}
	return NewResolver(name, resolverPort, resolverTlsCertPath, resolverTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath, errorLogPath, brokers, topics)
}
