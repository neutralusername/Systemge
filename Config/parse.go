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

	var tcpTimeoutMs uint64 = 5000
	infoPath := ""
	warningPath := ""
	errorPath := ""
	debugPath := ""
	var syncResponseTimeoutMs uint64 = 10000
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
	smtpHost := ""
	smtpPort := uint16(0)
	smtpUsername := ""
	smtpPassword := ""
	smtpRecipients := []string{}

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
		lineSegmentss := strings.Split(line, "\t")
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
			tcpTimeoutMs = Utilities.StringToUint64(lineSegments[1])
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
		case "_smtp":
			if len(lineSegments) < 6 {
				panic("smtp line is invalid \"" + line + "\"")
			}
			smtpHost = lineSegments[1]
			smtpPort = Utilities.StringToUint16(lineSegments[2])
			smtpUsername = lineSegments[3]
			smtpPassword = lineSegments[4]
			smtpRecipients = lineSegments[5:]
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
			syncResponseTimeoutMs = Utilities.StringToUint64(lineSegments[1])
		case "_sync":
			syncTopics = append(syncTopics, lineSegments[1:]...)
		case "_async":
			asyncTopics = append(asyncTopics, lineSegments[1:]...)
		}
	}
	if brokerPort == "" || brokerCertPath == "" || brokerKeyPath == "" || configPort == "" || configCertPath == "" || configKeyPath == "" {
		panic("missing required fields in config")
	}
	port := Utilities.StringToUint16(brokerPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + brokerPort + "\"")
	}
	brokerServer := TcpServer.New(port, brokerCertPath, brokerKeyPath)
	port = Utilities.StringToUint16(configPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + configPort + "\"")
	}
	configServer := TcpServer.New(port, configCertPath, configKeyPath)
	brokerEndpoint := TcpEndpoint.New(publicIp+":"+brokerPort, serverNameIndication, Utilities.GetFileContent(brokerCertPath))
	resolverConfigEndpoint := TcpEndpoint.New(resolverConfigAddress, resolverConfigServerNameIndication, Utilities.GetFileContent(resolverConfigCertPath))
	var mailer *Utilities.Mailer = nil
	if smtpHost != "" && smtpPort != 0 && smtpUsername != "" && smtpPassword != "" && len(smtpRecipients) > 0 {
		mailer = Utilities.NewMailer(smtpHost, smtpPort, smtpUsername, smtpPassword, smtpRecipients)
	}
	return Broker{
		Name:                   name,
		Logger:                 Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath, mailer),
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

	var tcpTimeoutMs uint64 = 5000
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
	smtpHost := ""
	smtpPort := uint16(0)
	smtpUsername := ""
	smtpPassword := ""
	smtpRecipients := []string{}

	lines := Utilities.SplitLines(Utilities.GetFileContent(sytemgeConfigPath))
	if len(lines) < 3 {
		panic("provided file has too few lines to be a valid config")
	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		lineSegmentss := strings.Split(line, "\t")
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
			tcpTimeoutMs = Utilities.StringToUint64(lineSegments[1])
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
		case "_smtp":
			if len(lineSegments) < 6 {
				panic("smtp line is invalid \"" + line + "\"")
			}
			smtpHost = lineSegments[1]
			smtpPort = Utilities.StringToUint16(lineSegments[2])
			smtpUsername = lineSegments[3]
			smtpPassword = lineSegments[4]
			smtpRecipients = lineSegments[5:]
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
	port := Utilities.StringToUint16(resolverPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + resolverPort + "\"")
	}
	resolverServer := TcpServer.New(port, resolverTlsCertPath, resolverTlsKeyPath)
	port = Utilities.StringToUint16(configPort)
	if port < 1 || port > 65535 {
		panic("invalid port \"" + configPort + "\"")
	}
	configServer := TcpServer.New(port, configTlsCertPath, configTlsKeyPath)
	var mailer *Utilities.Mailer = nil
	if smtpHost != "" && smtpPort != 0 && smtpUsername != "" && smtpPassword != "" && len(smtpRecipients) > 0 {
		mailer = Utilities.NewMailer(smtpHost, smtpPort, smtpUsername, smtpPassword, smtpRecipients)
	}
	return Resolver{
		Name:         name,
		Logger:       Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath, mailer),
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

	var tcpTimeoutMs uint64 = 5000
	var brokerSubscribeDelayMs uint64 = 1000
	var SyncResponseTimeoutMs uint64 = 10000
	var TopicResolutionLifetimeMs uint64 = 10000
	infoPath := ""
	warningPath := ""
	errorPath := ""
	debugPath := ""
	resolverAddress := ""
	resolverServerNameIndication := ""
	resolverCertPath := ""
	smtpHost := ""
	smtpPort := uint16(0)
	smtpUsername := ""
	smtpPassword := ""
	smtpRecipients := []string{}

	lines := Utilities.SplitLines(Utilities.GetFileContent(sytemgeConfigPath))
	if len(lines) < 2 {
		panic("provided file has too few lines to be a valid config")
	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		lineSegmentss := strings.Split(line, "\t")
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
			tcpTimeoutMs = Utilities.StringToUint64(lineSegments[1])
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
		case "_smtp":
			if len(lineSegments) < 6 {
				panic("smtp line is invalid \"" + line + "\"")
			}
			smtpHost = lineSegments[1]
			smtpPort = Utilities.StringToUint16(lineSegments[2])
			smtpUsername = lineSegments[3]
			smtpPassword = lineSegments[4]
			smtpRecipients = lineSegments[5:]
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
			brokerSubscribeDelayMs = Utilities.StringToUint64(lineSegments[1])
		case "_syncResponseTimeoutMs":
			if len(lineSegments) != 2 {
				panic("syncResponseTimeoutMs line is invalid \"" + line + "\"")
			}
			SyncResponseTimeoutMs = Utilities.StringToUint64(lineSegments[1])
		case "_topicResolutionLifetimeMs":
			if len(lineSegments) != 2 {
				panic("topicResolutionLifetimeMs line is invalid \"" + line + "\"")
			}
			TopicResolutionLifetimeMs = Utilities.StringToUint64(lineSegments[1])
		}
	}
	if resolverAddress == "" || resolverCertPath == "" {
		panic("missing required fields in config")
	}
	var mailer *Utilities.Mailer = nil
	if smtpHost != "" && smtpPort != 0 && smtpUsername != "" && smtpPassword != "" && len(smtpRecipients) > 0 {
		mailer = Utilities.NewMailer(smtpHost, smtpPort, smtpUsername, smtpPassword, smtpRecipients)
	}
	resolverEndpoint := TcpEndpoint.New(resolverAddress, resolverServerNameIndication, Utilities.GetFileContent(resolverCertPath))
	return Node{
		Name:                      name,
		Logger:                    Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath, mailer),
		ResolverEndpoint:          resolverEndpoint,
		TcpTimeoutMs:              tcpTimeoutMs,
		BrokerSubscribeDelayMs:    brokerSubscribeDelayMs,
		SyncResponseTimeoutMs:     SyncResponseTimeoutMs,
		TopicResolutionLifetimeMs: TopicResolutionLifetimeMs,
	}
}
