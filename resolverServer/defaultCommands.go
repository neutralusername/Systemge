package resolverServer

import (
	"encoding/json"
	"errors"

	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

func (server *Resolver[B, O]) GetDefaultCommands() tools.CommandHandlers {
	commands := tools.CommandHandlers{
		"start": func(args []string) (string, error) {
			err := server.singleRequestServerSync.GetRoutine().StartRoutine()
			if err != nil {
				return "", err
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := server.singleRequestServerSync.GetRoutine().StopRoutine(true)
			if err != nil {
				return "", err
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return status.ToString(server.singleRequestServerSync.GetRoutine().GetStatus()), nil
		},
		"checkMetrics": func(args []string) (string, error) {
			metrics := server.CheckMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", err
			}
			return string(json), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := server.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", err
			}
			return string(json), nil
		},
		/* "addResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", nil
			}
			tcpClientConfig := configs.UnmarshalTcpClient(args[1])
			if tcpClientConfig == nil {
				return "", nil
			}
			server.AddResolution(args[0], tcpClientConfig)
			return "success", nil
		}, */
		"removeResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", errors.New("expected 1 argument")
			}
			server.RemnoveResolution(args[0])
			return "success", nil
		},
	}
	systemgeServerCommands := server.singleRequestServerSync.GetDefaultCommands()
	for key, value := range systemgeServerCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
