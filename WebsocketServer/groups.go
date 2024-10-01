package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

func (server *WebsocketServer) GetGroupWebsocketConnectionIds(groupId string) ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Context{
				Event.Circumstance: Event.GetGroupClients,
				Event.IdentityType: Event.WebsocketConnection,
				Event.GroupId:      groupId,
			},
		))
		return nil, errors.New("group does not exist")
	}

	groupClientIds := make([]string, 0)
	for websocketId := range server.groupsWebsocketConnections[groupId] {
		groupClientIds = append(groupClientIds, websocketId)
	}

	return groupClientIds, nil
}

func (server *WebsocketServer) GetWebsocketConnectionGroupIds(websocketId string) ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if server.websocketConnections[websocketId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			Event.Context{
				Event.Circumstance: Event.GetClientGroups,
				Event.IdentityType: Event.WebsocketConnection,
				Event.ClientId:     websocketId,
			},
		))
		return nil, errors.New("websocketConnection does not exist")
	}

	groupIds := make([]string, 0)
	for groupId := range server.websocketConnectionGroups[websocketId] {
		groupIds = append(groupIds, groupId)
	}

	return groupIds, nil
}

// GetGroupCount returns the number of groups.
func (server *WebsocketServer) GetGroupCount() (int, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()
	groupCount := len(server.groupsWebsocketConnections)
	return groupCount, nil
}

func (server *WebsocketServer) GetGroupIds() ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	groups := make([]string, 0)
	for groupId := range server.groupsWebsocketConnections {
		groups = append(groups, groupId)
	}

	return groups, nil
}

func (server *WebsocketServer) IsWebsocketConnectionInGroup(groupId string, websocketId string) (bool, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Context{
				Event.Circumstance: Event.IsClientInGroup,
				Event.IdentityType: Event.WebsocketConnection,
				Event.ClientId:     websocketId,
				Event.GroupId:      groupId,
				Event.Result:       Helpers.JsonMarshal(false),
			},
		))
		return false, errors.New("group does not exist")
	}

	return true, nil
}

func (server *WebsocketServer) GetWebsocketConnectionGroupCount(websocketId string) (int, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if server.websocketConnectionGroups[websocketId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection not in any group",
			Event.Context{
				Event.Circumstance: Event.ClientGroupCount,
				Event.IdentityType: Event.WebsocketConnection,
				Event.ClientId:     websocketId,
			},
		))
		return -1, errors.New("websocketConnection not in any group")
	}

	groupCount := len(server.websocketConnectionGroups[websocketId])

	return groupCount, nil
}
