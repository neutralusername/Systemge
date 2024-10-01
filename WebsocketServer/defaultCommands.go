package WebsocketServer

import (
	"encoding/json"
	"errors"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
)

func (server *WebsocketServer) GetDefaultCommands() Commands.Handlers {
	httpCommands := server.httpServer.GetDefaultCommands()
	commands := Commands.Handlers{}
	for key, value := range httpCommands {
		commands["http_"+key] = value
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
	commands["broadcast"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", errors.New("invalid number of arguments")
		}
		topic := args[0]
		payload := args[1]
		err := server.Broadcast(Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["unicast"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", errors.New("invalid number of arguments")
		}
		topic := args[0]
		payload := args[1]
		id := args[2]
		err := server.Unicast(id, Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["multicast"] = func(args []string) (string, error) {
		if len(args) < 3 {
			return "", errors.New("invalid number of arguments")
		}
		topic := args[0]
		payload := args[1]
		ids := args[2:]
		err := server.Multicast(ids, Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	return commands
}
