package BrokerResolver

import (
	"encoding/json"

	"github.com/neutralusername/systemge/Commands"
	"github.com/neutralusername/systemge/Config"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/status"
)

func (server *Resolver) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{
		"start": func(args []string) (string, error) {
			err := server.Start()
			if err != nil {
				return "", Event.New("failed to start message broker server", err)
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := server.Stop()
			if err != nil {
				return "", Event.New("failed to stop message broker server", err)
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return status.ToString(server.GetStatus()), nil
		},
		"checkMetrics": func(args []string) (string, error) {
			metrics := server.CheckMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Event.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := server.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Event.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"addAsyncResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Event.New("expected 2 arguments", nil)
			}
			tcpClientConfig := Config.UnmarshalTcpClient(args[1])
			if tcpClientConfig == nil {
				return "", Event.New("failed unmarshalling tcpClientConfig", nil)
			}
			server.AddAsyncResolution(args[0], tcpClientConfig)
			return "success", nil
		},
		"removeAsyncResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Event.New("expected 1 argument", nil)
			}
			server.RemoveAsyncResolution(args[0])
			return "success", nil
		},
		"addSyncResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Event.New("expected 2 arguments", nil)
			}
			tcpClientConfig := Config.UnmarshalTcpClient(args[1])
			if tcpClientConfig == nil {
				return "", Event.New("failed unmarshalling tcpClientConfig", nil)
			}
			server.AddSyncResolution(args[0], tcpClientConfig)
			return "success", nil
		},
		"removeSyncResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Event.New("expected 1 argument", nil)
			}
			server.RemoveSyncResolution(args[0])
			return "success", nil
		},
	}
	systemgeServerCommands := server.systemgeServer.GetDefaultCommands()
	for key, value := range systemgeServerCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
