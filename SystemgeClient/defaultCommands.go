package SystemgeClient

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
)

func (client *SystemgeClient) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	commands["start"] = func(args []string) (string, error) {
		if err := client.Start(); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		if err := client.Stop(); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(client.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		metrics := client.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := client.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["addConnectionAttempt"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		tcpClientConfig := Config.UnmarshalTcpClient(args[0])
		if tcpClientConfig == nil {
			return "", Error.New("failed unmarshalling tcpClientConfig", nil)
		}
		if err := client.AddConnectionAttempt(tcpClientConfig); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["removeConnection"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		if err := client.RemoveConnection(args[0]); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getConnectionNamesAndAddresses"] = func(args []string) (string, error) {
		connectionNamesAndAddress := client.GetConnectionNamesAndAddresses()
		json, err := json.Marshal(connectionNamesAndAddress)
		if err != nil {
			return "", Error.New("failed to marshal connectionNamesAndAddress to json", err)
		}
		return string(json), nil
	}
	commands["getConnectionName"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		connectionName := client.GetConnectionName(args[0])
		if connectionName == "" {
			return "", Error.New("failed to get connection name", nil)
		}
		return connectionName, nil
	}
	commands["getConnectionAddress"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		connectionAddress := client.GetConnectionAddress(args[0])
		if connectionAddress == "" {
			return "", Error.New("failed to get connection address", nil)
		}
		return connectionAddress, nil
	}
	commands["getConnectionCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(client.GetConnectionCount()), nil
	}
	commands["asyncMessage"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		if err := client.AsyncMessage(topic, payload, clientNames...); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["syncRequest"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		messages, err := client.SyncRequestBlocking(topic, payload, clientNames...)
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(messages)
		if err != nil {
			return "", Error.New("failed to marshal messages to json", err)
		}
		return string(json), nil
	}
	return commands
}
