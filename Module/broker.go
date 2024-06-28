package Module

import (
	"Systemge/Config"
	"Systemge/Utilities"
	"strings"
)

func ParseBrokerConfigFromFile(sytemgeConfigPath string) Config.Broker {
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
	errorLogPath := ""
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
		case "_logs":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			errorLogPath = lineSegments[1]
		case "_broker":
			if len(lineSegments) != 4 {
				panic("broker line is invalid \"" + line + "\"")
			}
			brokerPort = lineSegments[1]
			brokerCertPath = lineSegments[2]
			brokerKeyPath = lineSegments[3]
		case "_config":
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
	return Config.Broker{
		Name:              name,
		LoggerPath:        errorLogPath,
		BrokerPort:        brokerPort,
		BrokerTlsCertPath: brokerCertPath,
		BrokerTlsKeyPath:  brokerKeyPath,
		ConfigPort:        configPort,
		ConfigTlsCertPath: configCertPath,
		ConfigTlsKeyPath:  configKeyPath,
		SyncTopics:        syncTopics,
		AsyncTopics:       asyncTopics,
	}
}
