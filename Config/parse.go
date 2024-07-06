package Config

import (
	"Systemge/TcpEndpoint"
	"Systemge/TcpServer"
	"Systemge/Utilities"
	"strings"
)

func ParseBrokerConfigFromFile(sytemgeConfigPath string) Broker {
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

	loggerPath := ""
	deliverImmediately := true
	nodeTimeoutMs := 10000
	syncRequestTimeoutMs := 10000
	publicIp := ""
	serverNameIndication := ""
	brokerPort := ""
	brokerCertPath := ""
	brokerKeyPath := ""
	configPort := ""
	configCertPath := ""
	configKeyPath := ""
	resolverAddress := ""
	resolverServerNameIndication := ""
	resolverCertPath := ""

	syncTopics := []string{}
	asyncTopics := []string{}
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
			loggerPath = lineSegments[1]
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
		case "_endpoint":
			if len(lineSegments) != 3 {
				panic("endpoint line is invalid \"" + line + "\"")
			}
			publicIp = lineSegments[1]
			serverNameIndication = lineSegments[2]
		case "_resolver":
			if len(lineSegments) != 4 {
				panic("resolver line is invalid \"" + line + "\"")
			}
			resolverAddress = lineSegments[1]
			resolverServerNameIndication = lineSegments[2]
			resolverCertPath = lineSegments[3]
		case "_deliverImmediately":
			if len(lineSegments) != 2 {
				panic("deliverImmediately line is invalid \"" + line + "\"")
			}
			deliverImmediately = lineSegments[1] == "true"
		case "_nodeTimeoutMs":
			if len(lineSegments) != 2 {
				panic("nodeTimeoutMs line is invalid \"" + line + "\"")
			}
			nodeTimeoutMs = Utilities.StringToInt(lineSegments[1])
		case "_syncRequestTimeoutMs":
			if len(lineSegments) != 2 {
				panic("syncRequestTimeoutMs line is invalid \"" + line + "\"")
			}
			syncRequestTimeoutMs = Utilities.StringToInt(lineSegments[1])
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
	if loggerPath == "" || brokerPort == "" || brokerCertPath == "" || brokerKeyPath == "" || configPort == "" || configCertPath == "" || configKeyPath == "" {
		panic("missing required fields in config")
	}
	port := Utilities.StringToInt(brokerPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + brokerPort + "\"")
	}
	brokerServer := TcpServer.New(port, brokerCertPath, brokerKeyPath)
	port = Utilities.StringToInt(configPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + configPort + "\"")
	}
	configServer := TcpServer.New(port, configCertPath, configKeyPath)
	brokerEndpoint := TcpEndpoint.New(publicIp+":"+brokerPort, serverNameIndication, Utilities.GetFileContent(brokerCertPath))
	resolverConfigEndpoint := TcpEndpoint.New(resolverAddress, resolverServerNameIndication, Utilities.GetFileContent(resolverCertPath))
	return Broker{
		Name:                   name,
		LoggerPath:             loggerPath,
		ResolverConfigEndpoint: resolverConfigEndpoint,
		DeliverImmediately:     deliverImmediately,
		NodeTimeoutMs:          nodeTimeoutMs,
		SyncRequestTimeoutMs:   syncRequestTimeoutMs,
		Server:                 brokerServer,
		Endpoint:               brokerEndpoint,
		ConfigServer:           configServer,
		SyncTopics:             syncTopics,
		AsyncTopics:            asyncTopics,
	}
}

func ParseResolverConfigFromFile(sytemgeConfigPath string) Resolver {
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

	loggerPath := ""
	resolverPort := ""
	resolverTlsCertPath := ""
	resolverTlsKeyPath := ""
	configPort := ""
	configTlsCertPath := ""
	configTlsKeyPath := ""
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
		case "_logs":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			loggerPath = lineSegments[1]
		case "_resolver":
			if len(lineSegments) != 4 {
				panic("resolver line is invalid \"" + line + "\"")
			}
			resolverPort = lineSegments[1]
			resolverTlsCertPath = lineSegments[2]
			resolverTlsKeyPath = lineSegments[3]
		case "_config":
			if len(lineSegments) != 4 {
				panic("config line is invalid \"" + line + "\"")
			}
			configPort = lineSegments[1]
			configTlsCertPath = lineSegments[2]
			configTlsKeyPath = lineSegments[3]
		}
	}
	if loggerPath == "" || resolverPort == "" || resolverTlsCertPath == "" || resolverTlsKeyPath == "" || configPort == "" || configTlsCertPath == "" || configTlsKeyPath == "" {
		panic("missing required fields in config")
	}
	port := Utilities.StringToInt(resolverPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + resolverPort + "\"")
	}
	resolverServer := TcpServer.New(port, resolverTlsCertPath, resolverTlsKeyPath)
	port = Utilities.StringToInt(configPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + configPort + "\"")
	}
	configServer := TcpServer.New(port, configTlsCertPath, configTlsKeyPath)
	return Resolver{
		Name:         name,
		LoggerPath:   loggerPath,
		Server:       resolverServer,
		ConfigServer: configServer,
	}
}
