package Config

import (
	"Systemge/TcpServer"
	"Systemge/Utilities"
	"strings"
)

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
		mailer = Utilities.NewMailer(smtpHost, smtpPort, smtpUsername, smtpPassword, smtpRecipients, Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath, nil))
	}
	return Resolver{
		Name:         name,
		Logger:       Utilities.NewLogger(infoPath, warningPath, errorPath, debugPath, mailer),
		Server:       resolverServer,
		ConfigServer: configServer,
		TcpTimeoutMs: tcpTimeoutMs,
	}
}
