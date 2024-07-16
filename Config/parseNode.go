package Config

import (
	"Systemge/TcpEndpoint"
	"Systemge/Utilities"
	"strings"
)

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
		mailer = Utilities.NewMailer(smtpHost, smtpPort, smtpUsername, smtpPassword, smtpRecipients, Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath, nil))
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
