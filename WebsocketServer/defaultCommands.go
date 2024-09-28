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
	commands["groupcast"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", errors.New("invalid number of arguments")
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
	commands["addWebsocketConnectionssToGroup_transactional"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.AddWebsocketConnectionsToGroup_transactional(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["addWebsocketConnectionsToGroup_bestEffort"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.AddWebsocketConnectionsToGroup_bestEffort(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["removeWebsocketConnectionsFromGroup_transactional"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.RemoveWebsocketConnectionsFromGroup_transactional(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["removeWebsocketConnectionsFromGroup_bestEffort"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", errors.New("invalid number of arguments")
		}
		group := args[0]
		ids := args[1:]
		err := server.RemoveWebsocketConnectionsFromGroup_bestEffort(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getGroupWebsocketConnectionIds"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("invalid number of arguments")
		}
		group := args[0]
		websocketConnectionIds, err := server.GetGroupWebsocketConnectionIds(group)
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(websocketConnectionIds)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getWebsocketConnectionGroupIds"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("invalid number of arguments")
		}
		id := args[0]
		groups, err := server.GetWebsocketConnectionGroupIds(id)
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
	commands["isWebsocketConnectionInGroup"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", errors.New("invalid number of arguments")
		}
		id := args[0]
		group := args[1]
		inGroup, err := server.IsWebsocketConnectionInGroup(id, group)
		if err != nil {
			return "", err
		}
		return Helpers.BoolToString(inGroup), nil
	}
	commands["websocketConnectionExists"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("invalid number of arguments")
		}
		websocketConnectionExists := server.WebsocketConnectionExists(args[0])
		return Helpers.BoolToString(websocketConnectionExists), nil
	}
	commands["getWebsocketConnectionGroupCount"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("invalid number of arguments")
		}
		id := args[0]
		websocketConnectionGroupCount, err := server.GetWebsocketConnectionGroupCount(id)
		if err != nil {
			return "", err
		}
		return Helpers.IntToString(websocketConnectionGroupCount), nil
	}
	commands["getWebsocketConnectionCount"] = func(args []string) (string, error) {
		websocketConnectionCount := server.GetWebsocketConnectionCount()
		return Helpers.IntToString(websocketConnectionCount), nil
	}
	commands["closeWebsocketConnection"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", errors.New("invalid number of arguments")
		}
		id := args[0]
		server.websocketConnectionMutex.Lock()
		defer server.websocketConnectionMutex.Unlock()
		if websocketConnection, ok := server.websocketConnections[id]; ok {
			websocketConnection.Close()
			return "success", nil
		}
		return "", errors.New("websocketConnection does not exist")
	}
	return commands
}
