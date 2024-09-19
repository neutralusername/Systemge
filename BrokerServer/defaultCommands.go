package BrokerServer

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) GetDefaultCommands() Commands.Handlers {
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
		"addAsyncTopics": func(args []string) (string, error) {
			server.AddAsyncTopics(args)
			return "success", nil
		},
		"removeAsyncTopics": func(args []string) (string, error) {
			server.RemoveAsyncTopics(args)
			return "success", nil
		},
		"addSyncTopics": func(args []string) (string, error) {
			server.AddSyncTopics(args)
			return "success", nil
		},
		"removeSyncTopics": func(args []string) (string, error) {
			server.RemoveSyncTopics(args)
			return "success", nil
		},
	}
	systemgeServerCommands := server.systemgeServer.GetDefaultCommands()
	for key, value := range systemgeServerCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
