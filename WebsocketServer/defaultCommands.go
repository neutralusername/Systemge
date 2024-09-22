package WebsocketServer

import (
	"encoding/json"
	"errors"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Helpers"
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
			return "", errors.New("Invalid number of arguments")
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
			return "", errors.New("Invalid number of arguments")
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
			return "", errors.New("Invalid number of arguments")
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
	commands["groupcast"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", errors.New("Invalid number of arguments")
		}
		topic := args[0]
		payload := args[1]
		group := args[2]
		err := server.Groupcast(group, Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["addClientsToGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("Invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.AddClientsToGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["attemptToAddClientsToGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("Invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.AttemptToAddClientsToGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["removeClientsFromGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("Invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.RemoveClientsFromGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["attemptToRemoveClientsFromGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("Invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.AttemptToRemoveClientsFromGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getGroupClients"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("Invalid number of arguments")
		}
		group := args[0]
		clients, err := server.GetGroupClients(group)
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(clients)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getClientGroups"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("Invalid number of arguments")
		}
		id := args[0]
		groups, err := server.GetClientGroups(id)
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(groups)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getGroupCount"] = func(args []string) (string, error) {
		groupCount, err := server.GetGroupCount()
		if err != nil {
			return "", err
		}
		return Helpers.IntToString(groupCount), nil
	}
	commands["getGroupIds"] = func(args []string) (string, error) {
		groups, err := server.GetGroupIds()
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(groups)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["isClientInGroup"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", errors.New("Invalid number of arguments")
		}
		id := args[0]
		group := args[1]
		inGroup, err := server.IsClientInGroup(id, group)
		if err != nil {
			return "", err
		}
		return Helpers.BoolToString(inGroup), nil
	}
	commands["clientExists"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("Invalid number of arguments")
		}
		id := args[0]
		clientExists, err := server.ClientExists(id)
		if err != nil {
			return "", err
		}
		return Helpers.BoolToString(clientExists), nil
	}
	commands["getClientGroupCount"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("Invalid number of arguments")
		}
		id := args[0]
		clientGroupCount, err := server.GetClientGroupCount(id)
		if err != nil {
			return "", err
		}
		return Helpers.IntToString(clientGroupCount), nil
	}
	commands["getClientCount"] = func(args []string) (string, error) {
		clientGroupCount, err := server.GetClientCount()
		if err != nil {
			return "", err
		}
		return Helpers.IntToString(clientGroupCount), nil
	}
	commands["closeClient"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("Invalid number of arguments")
		}
		id := args[0]
		server.clientMutex.Lock()
		defer server.clientMutex.Unlock()
		if client, ok := server.clients[id]; ok {
			client.Close()
			return "success", nil
		}
		return "", errors.New("client does not exist")
	}
	return commands
}
