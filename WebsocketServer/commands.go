package WebsocketServer

import (
	"encoding/json"
	"errors"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Status"
)

func (server *WebsocketServer[T]) GetDefaultCommands() Commands.Handlers {
	websocketListenerCommands := server.websocketListener.GetDefaultCommands()
	commands := Commands.Handlers{}
	for key, value := range websocketListenerCommands {
		commands["websocketListener_"+key] = value
	}
	commands["start"] = func(args []string) (string, error) {
		err := server.Start()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		err := server.Stop()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(server.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		metrics := server.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := server.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["asyncMessage"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("invalid number of arguments")
		}
		bytes, err := json.Marshal(args[0])
		if err != nil {
			return "", err
		}
		ids := args[1:]
		err = server.AsyncMessage(bytes, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	/* 	commands["syncMessage"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("invalid number of arguments")
		}
		topic := args[0]
		payload := args[1]
		syncToken := args[2]
		ids := args[3:]
		responses, err := server.SyncRequestBlocking(topic, payload, syncToken, ids...)
		if err != nil {
			return "", err
		}
		return Helpers.JsonMarshal(responses), nil
	} */
	return commands
}
