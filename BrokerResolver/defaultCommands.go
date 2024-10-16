package BrokerResolver

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Resolver) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{
		"start": func(args []string) (string, error) {
			err := server.Start()
			if err != nil {
				return "", Error.New("failed to start message broker server", err)
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := server.Stop()
			if err != nil {
				return "", Error.New("failed to stop message broker server", err)
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return Status.ToString(server.GetStatus()), nil
		},
		"checkMetrics": func(args []string) (string, error) {
			metrics := server.CheckMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := server.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"addAsyncResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Error.New("expected 2 arguments", nil)
			}
			tcpClientConfig := Config.UnmarshalTcpClient(args[1])
			if tcpClientConfig == nil {
				return "", Error.New("failed unmarshalling tcpClientConfig", nil)
			}
			server.AddAsyncResolution(args[0], tcpClientConfig)
			return "success", nil
		},
		"removeAsyncResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("expected 1 argument", nil)
			}
			server.RemoveAsyncResolution(args[0])
			return "success", nil
		},
		"addSyncResolution": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Error.New("expected 2 arguments", nil)
			}
			tcpClientConfig := Config.UnmarshalTcpClient(args[1])
			if tcpClientConfig == nil {
				return "", Error.New("failed unmarshalling tcpClientConfig", nil)
			}
			server.AddSyncResolution(args[0], tcpClientConfig)
			return "success", nil
		},
		"removeSyncResolution": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("expected 1 argument", nil)
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
