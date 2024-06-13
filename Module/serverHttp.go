package Module

import (
	"Systemge/HTTP"
	"Systemge/Utilities"
	"fmt"
	"strings"
)

func NewHTTPServer(name string, port string, loggerPath string, tlsCert string, tlsKey string, patterns map[string]HTTP.RequestHandler) *HTTP.Server {
	logger := Utilities.NewLogger(loggerPath)
	httpServer := HTTP.New(port, name, tlsCert, tlsKey, logger)
	for pattern, handler := range patterns {
		httpServer.RegisterPattern(pattern, handler)
	}
	return httpServer
}

func NewHTTPServerFromConfig(sytemgeConfigPath string, errorLogPath string) *HTTP.Server {
	if !Utilities.FileExists(sytemgeConfigPath) {
		panic("provided file does not exist")
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
		panic("error reading file")
	}
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		switch i {
		case 0:
			if len(line) < 2 {
				panic("error reading file. missing config type")
			}
			segments := strings.Split(line, " ")
			if len(segments) != 2 {
				panic("error reading file. incomplete config type")
			}
			if segments[0] != "#" {
				panic("error reading file. missing config type identifier (#)")
			}
			if segments[1] != "http" {
				panic("error reading file. invalid config type (want: http)")
			}
			continue
		case 1:
			if len(line) < 2 {
				panic("error reading file. Missing port number")
			}
			if line[0] != ':' {
				panic("error reading file. Missing port number")
			}
			port = line
		default:
			segments := strings.Split(line, " ")
			if len(segments) == 3 {
				pattern := segments[0]
				handlerType := segments[1]
				handlerArg := segments[2]
				switch handlerType {
				case "serve":
					patterns[pattern] = HTTP.SendDirectory(handlerArg)
				case "redirect":
					patterns[pattern] = HTTP.RedirectTo(handlerArg)
				default:
					panic("error reading file. incorrect handler type")
				}
			} else if len(segments) == 2 {
				switch segments[0] {
				case "key":
					tlsKeyPath = segments[1]
				case "cert":
					tlsCertPath = segments[1]
				default:
					panic("error reading file. incorrect row or cert/key")
				}
			} else {
				println(len(segments))
				fmt.Println(segments)
				panic("error reading file. incorrect row or cert/key")
			}
		}
	}
	return NewHTTPServer(name, port, errorLogPath, tlsCertPath, tlsKeyPath, patterns)
}
