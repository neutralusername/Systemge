package WebsocketServer

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Error"
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
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		err := server.Broadcast(Message.New(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["unicast"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		id := args[2]
		err := server.Unicast(id, Message.New(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["multicast"] = func(args []string) (string, error) {
		if len(args) < 3 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		ids := args[2:]
		err := server.Multicast(ids, Message.New(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["groupcast"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		group := args[2]
		err := server.Groupcast(group, Message.New(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["addClientsToGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("Invalid number of arguments", nil)
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
			return "", Error.New("Invalid number of arguments", nil)
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
			return "", Error.New("Invalid number of arguments", nil)
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
			return "", Error.New("Invalid number of arguments", nil)
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
			return "", Error.New("Invalid number of arguments", nil)
		}
		group := args[0]
		clients := server.GetGroupClients(group)
		json, err := json.Marshal(clients)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getClientGroups"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		groups := server.GetClientGroups(id)
		json, err := json.Marshal(groups)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getGroupCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(server.GetGroupCount()), nil
	}
	commands["getGroups"] = func(args []string) (string, error) {
		groups := server.GetGroups()
		json, err := json.Marshal(groups)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["isClientInGroup"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		group := args[1]
		return Helpers.BoolToString(server.IsClientInGroup(id, group)), nil
	}
	commands["clientExists"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		return Helpers.BoolToString(server.ClientExists(id)), nil
	}
	commands["getClientGroupCount"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		return Helpers.IntToString(server.GetClientGroupCount(id)), nil
	}
	commands["getClientCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(server.GetClientCount()), nil
	}
	commands["resetClientWatchdog"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		server.clientMutex.Lock()
		defer server.clientMutex.Unlock()
		if client, ok := server.clients[id]; ok {
			server.ResetWatchdog(client)
			return "success", nil
		}
		return "", Error.New("client with id "+id+" does not exist", nil)
	}
	commands["disconnectClient"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		server.clientMutex.Lock()
		defer server.clientMutex.Unlock()
		if client, ok := server.clients[id]; ok {
			client.Disconnect()
			return "success", nil
		}
		return "", Error.New("client with id "+id+" does not exist", nil)
	}
	return commands
}
