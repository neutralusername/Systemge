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

	tcpTimeoutMs := 5000
	infoPath := ""
	warningPath := ""
	errorPath := ""
	debugPath := ""
	syncResponseTimeoutMs := 10000
	publicIp := ""
	serverNameIndication := ""
	brokerPort := ""
	brokerCertPath := ""
	brokerKeyPath := ""
	configPort := ""
	configCertPath := ""
	configKeyPath := ""
	resolverConfigAddress := ""
	resolverConfigServerNameIndication := ""
	resolverConfigCertPath := ""

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
		case "_tcpTimeoutMs":
			if len(lineSegments) != 2 {
				panic("tcpTimeoutMs line is invalid \"" + line + "\"")
			}
			tcpTimeoutMs = Utilities.StringToInt(lineSegments[1])
		case "_info":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			infoPath = lineSegments[1]
		case "_warning":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			warningPath = lineSegments[1]
		case "_error":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			errorPath = lineSegments[1]
		case "_debug":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			debugPath = lineSegments[1]
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
		case "_resolverConfig":
			if len(lineSegments) != 4 {
				panic("resolver line is invalid \"" + line + "\"")
			}
			resolverConfigAddress = lineSegments[1]
			resolverConfigServerNameIndication = lineSegments[2]
			resolverConfigCertPath = lineSegments[3]
		case "_syncResponseTimeoutMs":
			if len(lineSegments) != 2 {
				panic("syncRequestTimeoutMs line is invalid \"" + line + "\"")
			}
			syncResponseTimeoutMs = Utilities.StringToInt(lineSegments[1])
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
	if brokerPort == "" || brokerCertPath == "" || brokerKeyPath == "" || configPort == "" || configCertPath == "" || configKeyPath == "" {
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
	resolverConfigEndpoint := TcpEndpoint.New(resolverConfigAddress, resolverConfigServerNameIndication, Utilities.GetFileContent(resolverConfigCertPath))
	return Broker{
		Name:                   name,
		Logger:                 Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath),
		ResolverConfigEndpoint: resolverConfigEndpoint,
		SyncResponseTimeoutMs:  syncResponseTimeoutMs,
		TcpTimeoutMs:           tcpTimeoutMs,
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

	tcpTimeoutMs := 5000
	infoPath := ""
	warningPath := ""
	errorPath := ""
	debugPath := ""
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
		case "_tcpTimeoutMs":
			if len(lineSegments) != 2 {
				panic("tcpTimeoutMs line is invalid \"" + line + "\"")
			}
			tcpTimeoutMs = Utilities.StringToInt(lineSegments[1])
		case "_info":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			infoPath = lineSegments[1]
		case "_warning":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			warningPath = lineSegments[1]
		case "_error":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			errorPath = lineSegments[1]
		case "_debug":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			debugPath = lineSegments[1]
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
	if resolverPort == "" || resolverTlsCertPath == "" || resolverTlsKeyPath == "" || configPort == "" || configTlsCertPath == "" || configTlsKeyPath == "" {
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
		Logger:       Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath),
		Server:       resolverServer,
		ConfigServer: configServer,
		TcpTimeoutMs: tcpTimeoutMs,
	}
}

func ParseNodeConfigFromFile(sytemgeConfigPath string) Node {
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

	tcpTimeoutMs := 5000
	brokerSubscribeDelayMs := 1000
	SyncResponseTimeoutMs := 10000
	TopicResolutionLifetimeMs := 10000
	infoPath := ""
	warningPath := ""
	errorPath := ""
	debugPath := ""
	resolverAddress := ""
	resolverServerNameIndication := ""
	resolverCertPath := ""

	lines := Utilities.SplitLines(Utilities.GetFileContent(sytemgeConfigPath))
	if len(lines) < 2 {
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
			if lineSegments[1] != "node" {
				panic("wrong config type for node \"" + lineSegments[1] + "\"")
			}
		case "_tcpTimeoutMs":
			if len(lineSegments) != 2 {
				panic("tcpTimeoutMs line is invalid \"" + line + "\"")
			}
			tcpTimeoutMs = Utilities.StringToInt(lineSegments[1])
		case "_info":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			infoPath = lineSegments[1]
		case "_warning":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			warningPath = lineSegments[1]
		case "_error":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			errorPath = lineSegments[1]
		case "_debug":
			if len(lineSegments) != 2 {
				panic("logs line is invalid \"" + line + "\"")
			}
			debugPath = lineSegments[1]
		case "_resolver":
			if len(lineSegments) != 4 {
				panic("resolver line is invalid \"" + line + "\"")
			}
			resolverAddress = lineSegments[1]
			resolverServerNameIndication = lineSegments[2]
			resolverCertPath = lineSegments[3]
		case "_brokerSubscribeDelayMs":
			if len(lineSegments) != 2 {
				panic("brokerSubscribeDelayMs line is invalid \"" + line + "\"")
			}
			brokerSubscribeDelayMs = Utilities.StringToInt(lineSegments[1])
		case "_syncResponseTimeoutMs":
			if len(lineSegments) != 2 {
				panic("syncResponseTimeoutMs line is invalid \"" + line + "\"")
			}
			SyncResponseTimeoutMs = Utilities.StringToInt(lineSegments[1])
		case "_topicResolutionLifetimeMs":
			if len(lineSegments) != 2 {
				panic("topicResolutionLifetimeMs line is invalid \"" + line + "\"")
			}
			TopicResolutionLifetimeMs = Utilities.StringToInt(lineSegments[1])
		}
	}
	if resolverAddress == "" || resolverCertPath == "" {
		panic("missing required fields in config")
	}
	resolverEndpoint := TcpEndpoint.New(resolverAddress, resolverServerNameIndication, Utilities.GetFileContent(resolverCertPath))
	return Node{
		Name:                      name,
		Logger:                    Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath),
		ResolverEndpoint:          resolverEndpoint,
		TcpTimeoutMs:              tcpTimeoutMs,
		BrokerSubscribeDelayMs:    brokerSubscribeDelayMs,
		SyncResponseTimeoutMs:     SyncResponseTimeoutMs,
		TopicResolutionLifetimeMs: TopicResolutionLifetimeMs,
	}
}
