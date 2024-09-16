package SystemgeServer

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
)

func (server *SystemgeServer) GetDefaultCommands() Commands.Handlers {
	serverCommands := Commands.Handlers{}
	if server.blacklist != nil {
		blacklistCommands := server.blacklist.GetDefaultCommands()
		for key, value := range blacklistCommands {
			serverCommands["blacklist_"+key] = value
		}
	}
	if server.whitelist != nil {
		whitelistCommands := server.whitelist.GetDefaultCommands()
		for key, value := range whitelistCommands {
			serverCommands["whitelist_"+key] = value
		}
	}
	serverCommands["start"] = func(args []string) (string, error) {
		err := server.Start()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["stop"] = func(args []string) (string, error) {
		err := server.Stop()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["removeConnection"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		err := server.RemoveConnection(args[0])
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["getConnectionNamesAndAddresses"] = func(args []string) (string, error) {
		names := server.GetConnectionNamesAndAddresses()
		json, err := json.Marshal(names)
		if err != nil {
			return "", Error.New("failed to marshal map to json", err)
		}
		return string(json), nil
	}
	serverCommands["getConnectionCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(server.GetConnectionCount()), nil
	}
	serverCommands["asyncMessage"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		err := server.AsyncMessage(topic, payload, clientNames...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	serverCommands["syncRequest"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		messages, err := server.SyncRequestBlocking(topic, payload, clientNames...)
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(messages)
		if err != nil {
			return "", Error.New("failed to marshal messages to json", err)
		}
		return string(json), nil
	}
	serverCommands["checkMetrics"] = func(args []string) (string, error) {
		metrics := server.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	serverCommands["getMetrics"] = func(args []string) (string, error) {
		metrics := server.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	return serverCommands
}
