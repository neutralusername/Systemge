package Module

import (
	"Systemge/HTTP"
	"Systemge/Utilities"
	"strings"
)

func NewHTTPServer(name string, port string, loggerPath string, tlsCertPath string, tlsKeyPath string, patterns map[string]HTTP.RequestHandler) *HTTP.Server {
	httpServer := HTTP.New(port, name, tlsCertPath, tlsKeyPath, Utilities.NewLogger(loggerPath))
	for pattern, handler := range patterns {
		httpServer.RegisterPattern(pattern, handler)
	}
	return httpServer
}

func NewHTTPServerFromConfig(sytemgeConfigPath string, errorLogPath string) *HTTP.Server {
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
	port := ""
	tlsCertPath := ""
	tlsKeyPath := ""
	patterns := map[string]HTTP.RequestHandler{}
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
			if lineSegments[1] != "http" {
				panic("wrong config type for http \"" + lineSegments[1] + "\"")
			}
		case "http":
			if len(lineSegments) != 2 {
				panic("http line is invalid \"" + line + "\"")
			}
			port = lineSegments[1]
		default:
			if len(lineSegments) != 3 {
				panic("handler line is invalid \"" + line + "\"")
			}
			if lineSegments[0][0] != '/' {
				panic("handler pattern is invalid \"" + lineSegments[0] + "\"")
			}
			switch lineSegments[1] {
			case "serve":
				patterns[lineSegments[0]] = HTTP.SendDirectory(lineSegments[2])
			case "redirect":
				patterns[lineSegments[0]] = HTTP.RedirectTo(lineSegments[2])
			default:
				panic("handler type is invalid \"" + lineSegments[1] + "\"")
			}
		}
	}
	return NewHTTPServer(name, port, errorLogPath, tlsCertPath, tlsKeyPath, patterns)
}
