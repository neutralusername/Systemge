package Module

import (
	"Systemge/Broker"
	"Systemge/Utilities"
	"strings"
)

func NewBroker(name, brokerPort, brokerTlsCertPath, brokerTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath, loggerPath string, asyncTopics []string, syncTopics []string) *Broker.Server {
	broker := Broker.New(name, brokerPort, brokerTlsCertPath, brokerTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath, Utilities.NewLogger(loggerPath))
	for _, topic := range asyncTopics {
		broker.AddAsyncTopics(topic)
	}
	for _, topic := range syncTopics {
		broker.AddSyncTopics(topic)
	}
	return broker
}

func NewBrokerFromConfig(sytemgeConfigPath string, errorLogPath string) *Broker.Server {
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
	brokerPort := ""
	brokerCertPath := ""
	brokerKeyPath := ""
	configPort := ""
	configCertPath := ""
	configKeyPath := ""
	asyncTopics := []string{}
	syncTopics := []string{}
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
			if lineSegments[1] != "broker" {
				panic("wrong config type for broker \"" + lineSegments[1] + "\"")
			}
		case "broker":
			if len(lineSegments) != 4 {
				panic("broker line is invalid \"" + line + "\"")
			}
			brokerPort = lineSegments[1]
			brokerCertPath = lineSegments[2]
			brokerKeyPath = lineSegments[3]
		case "config":
			if len(lineSegments) != 4 {
				panic("config line is invalid \"" + line + "\"")
			}
			configPort = lineSegments[1]
			configCertPath = lineSegments[2]
			configKeyPath = lineSegments[3]
		default:
			if len(lineSegments) != 2 {
				panic("invalid topic line \"" + line + "\"")
			}
			if lineSegments[1] == "sync" {
				syncTopics = append(syncTopics, lineSegments[0])
			} else if lineSegments[1] == "async" {
				asyncTopics = append(asyncTopics, lineSegments[0])
			} else {
				println(lineSegments[1])
				panic("invalid topic type \"" + lineSegments[1] + "\"")
			}
		}
	}
	return NewBroker(name, brokerPort, brokerCertPath, brokerKeyPath, configPort, configCertPath, configKeyPath, errorLogPath, asyncTopics, syncTopics)
}
