package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

func (server *WebsocketServer) GetGroupWebsocketConnectionIds(groupId string) ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingGroupClients,
		"getting group websocketConnections",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.GetGroupClientsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.GroupId:      groupId,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Context{
				Event.Circumstance: Event.GetGroupClientsRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.GroupId:      groupId,
			},
		))
		return nil, errors.New("group does not exist")
	}

	groupClientIds := make([]string, 0)
	for websocketId := range server.groupsWebsocketConnections[groupId] {
		groupClientIds = append(groupClientIds, websocketId)
	}

	if event := server.onEvent(Event.NewInfo(
		Event.GotGroupClients,
		"got group websocketConnections",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.GetGroupClientsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.GroupId:      groupId,
			Event.Result:       Helpers.JsonMarshal(groupClientIds),
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	return groupClientIds, nil
}

func (server *WebsocketServer) GetWebsocketConnectionGroupIds(websocketId string) ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingClientGroups,
		"getting websocketConnection groups",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.GetClientGroupsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	if server.websocketConnections[websocketId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			Event.Context{
				Event.Circumstance: Event.GetClientGroupsRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.ClientId:     websocketId,
			},
		))
		return nil, errors.New("websocketConnection does not exist")
	}

	groupIds := make([]string, 0)
	for groupId := range server.websocketConnectionGroups[websocketId] {
		groupIds = append(groupIds, groupId)
	}

	if event := server.onEvent(Event.NewInfo(
		Event.GotClientGroups,
		"got websocketConnection groups",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.GetClientGroupsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.Result:       Helpers.JsonMarshal(groupIds),
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}
	return groupIds, nil
}

// GetGroupCount returns the number of groups.
func (server *WebsocketServer) GetGroupCount() (int, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingGroupCount,
		"getting group count",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.GetGroupCountRoutine,
			Event.ClientType:   Event.WebsocketConnection,
		},
	)); !event.IsInfo() {
		return -1, event.GetError()
	}

	groupCount := len(server.groupsWebsocketConnections)

	if event := server.onEvent(Event.NewInfoNoOption(
		Event.GotGroupCount,
		"got group count",
		Event.Context{
			Event.Circumstance: Event.GetGroupCountRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.Result:       Helpers.JsonMarshal(groupCount),
		},
	)); !event.IsInfo() {
		return -1, event.GetError()
	}
	return groupCount, nil
}

func (server *WebsocketServer) GetGroupIds() ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingGroupIds,
		"getting group ids",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.GetGroupIdsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	groups := make([]string, 0)
	for groupId := range server.groupsWebsocketConnections {
		groups = append(groups, groupId)
	}

	if event := server.onEvent(Event.NewInfoNoOption(
		Event.GotGroupIds,
		"got group ids",
		Event.Context{
			Event.Circumstance: Event.GetGroupIdsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.Result:       Helpers.JsonMarshal(groups),
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}
	return groups, nil
}

func (server *WebsocketServer) IsWebsocketConnectionInGroup(groupId string, websocketId string) (bool, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingIsClientInGroup,
		"getting is websocketConnection in group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.IsClientInGroupRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.GroupId:      groupId,
		},
	)); !event.IsInfo() {
		return false, event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Context{
				Event.Circumstance: Event.IsClientInGroupRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.ClientId:     websocketId,
				Event.GroupId:      groupId,
				Event.Result:       Helpers.JsonMarshal(false),
			},
		))
		return false, errors.New("group does not exist")
	}

	if event := server.onEvent(Event.NewInfoNoOption(
		Event.GotIsClientInGroup,
		"got is websocketConnection in group",
		Event.Context{
			Event.Circumstance: Event.IsClientInGroupRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.GroupId:      groupId,
			Event.Result:       Helpers.JsonMarshal(true),
		},
	)); !event.IsInfo() {
		return false, event.GetError()
	}
	return true, nil
}

func (server *WebsocketServer) GetWebsocketConnectionGroupCount(websocketId string) (int, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingClientGroupCount,
		"getting websocketConnection group count",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ClientGroupCountRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
		},
	)); !event.IsInfo() {
		return -1, event.GetError()
	}

	if server.websocketConnectionGroups[websocketId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection not in any group",
			Event.Context{
				Event.Circumstance: Event.ClientGroupCountRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.ClientId:     websocketId,
			},
		))
		return -1, errors.New("websocketConnection not in any group")
	}

	groupCount := len(server.websocketConnectionGroups[websocketId])

	if event := server.onEvent(Event.NewInfo(
		Event.GotClientGroupCount,
		"got websocketConnection group count",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ClientGroupCountRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.Result:       Helpers.JsonMarshal(groupCount),
		},
	)); !event.IsInfo() {
		return -1, event.GetError()
	}
	return groupCount, nil
}
