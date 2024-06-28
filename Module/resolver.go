package Module

import (
	"Systemge/Config"
	"Systemge/Resolution"
	"Systemge/Utilities"
	"strings"
)

func ParseResolverConfigFromFile(sytemgeConfigPath string) Config.Resolver {
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
	errorLogPath := ""
	topics := map[string]Resolution.Resolution{} // topic -> broker
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
			errorLogPath = lineSegments[1]
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
		default:
			if len(lineSegments) < 4 {
				panic("broker line is invalid \"" + line + "\"")
			}
			if !Utilities.FileExists(lineSegments[3]) {
				panic("certificate file does not exist \"" + lineSegments[2] + "\"")
			}
			resolution := Resolution.New(lineSegments[0], lineSegments[1], lineSegments[2], Utilities.GetFileContent(lineSegments[3]))
			for _, topic := range lineSegments[3:] {
				topics[topic] = *resolution
			}
		}
	}
	return Config.Resolver{
		Name:                name,
		LoggerPath:          errorLogPath,
		ResolverPort:        resolverPort,
		ResolverTlsCertPath: resolverTlsCertPath,
		ResolverTlsKeyPath:  resolverTlsKeyPath,
		ConfigPort:          configPort,
		ConfigTlsCertPath:   configTlsCertPath,
		ConfigTlsKeyPath:    configTlsKeyPath,
		TopicResolutions:    topics,
	}
}
