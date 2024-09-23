package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

func (server *WebsocketServer) AddWebsocketConnectionsToGroup_transactional(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	targetClientIds := Helpers.JsonMarshal(websocketIds)
	if event := server.onEvent(Event.NewInfo(
		Event.AddingClientsToGroup,
		"adding websocketConnection to group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.AddClientsToGroupRoutine,
			Event.Behaviour:       Event.Transactional,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] == nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.AddClientsToGroupRoutine,
					Event.Behaviour:       Event.Transactional,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        websocketId,
					Event.TargetClientIds: targetClientIds,
					Event.GroupId:         groupId,
				}),
			))
			return errors.New("client does not exist")
		}
		if server.websocketConnectionGroups[websocketId][groupId] {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientAlreadyInGroup,
				"websocketConnection is already in group",
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.AddClientsToGroupRoutine,
					Event.Behaviour:       Event.Transactional,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        websocketId,
					Event.TargetClientIds: targetClientIds,
					Event.GroupId:         groupId,
				}),
			))
			return errors.New("websocketConnection is already in group")
		}
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		if event := server.onEvent(Event.NewInfo(
			Event.CreatingGroup,
			"creating group",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:    Event.AddClientsToGroupRoutine,
				Event.Behaviour:       Event.Transactional,
				Event.ClientType:      Event.WebsocketConnection,
				Event.TargetClientIds: targetClientIds,
				Event.GroupId:         groupId,
			}),
		)); !event.IsInfo() {
			return event.GetError()
		}
		server.groupsWebsocketConnections[groupId] = make(map[string]*WebsocketConnection)
	}

	for _, websocketId := range websocketIds {
		server.groupsWebsocketConnections[groupId][websocketId] = server.websocketConnections[websocketId]
		server.websocketConnectionGroups[websocketId][groupId] = true
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.AddedClientsToGroup,
		"added websocketConnections to group",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.AddClientsToGroupRoutine,
			Event.Behaviour:       Event.Transactional,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	))
	return nil
}

func (server *WebsocketServer) AddWebsocketConnectionsToGroup_bestEffort(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	targetClientIds := Helpers.JsonMarshal(websocketIds)
	if event := server.onEvent(Event.NewInfo(
		Event.AddingClientsToGroup,
		"adding websocketConnections to group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.AddClientsToGroupRoutine,
			Event.Behaviour:       Event.BestEffort,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		if event := server.onEvent(Event.NewWarning(
			Event.CreatingGroup,
			"creating group",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:    Event.AddClientsToGroupRoutine,
				Event.Behaviour:       Event.BestEffort,
				Event.ClientType:      Event.WebsocketConnection,
				Event.TargetClientIds: targetClientIds,
				Event.GroupId:         groupId,
			}),
		)); !event.IsInfo() {
			return event.GetError()
		}
		server.groupsWebsocketConnections[groupId] = make(map[string]*WebsocketConnection)
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] != nil {
			server.groupsWebsocketConnections[groupId][websocketId] = server.websocketConnections[websocketId]
			server.websocketConnectionGroups[websocketId][groupId] = true
		} else {
			event := server.onEvent(Event.NewWarning(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.AddClientsToGroupRoutine,
					Event.Behaviour:       Event.BestEffort,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        websocketId,
					Event.TargetClientIds: targetClientIds,
					Event.GroupId:         groupId,
				}),
			))
			if !event.IsInfo() {
				return event.GetError()
			}
		}
	}

	if len(server.groupsWebsocketConnections[groupId]) == 0 {
		delete(server.groupsWebsocketConnections, groupId)
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.AddedClientsToGroup,
		"added websocketConnections to group",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.AddClientsToGroupRoutine,
			Event.Behaviour:       Event.BestEffort,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	))
	return nil
}

func (server *WebsocketServer) RemoveWebsocketConnectionsFromGroup_transactional(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	targetClientIds := Helpers.JsonMarshal(websocketIds)
	if event := server.onEvent(Event.NewInfo(
		Event.RemovingClientsFromGroup,
		"removing websocketConnections from group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
			Event.Behaviour:       Event.Transactional,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
				Event.Behaviour:       Event.Transactional,
				Event.ClientType:      Event.WebsocketConnection,
				Event.TargetClientIds: targetClientIds,
				Event.GroupId:         groupId,
			}),
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] == nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
					Event.Behaviour:       Event.Transactional,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        websocketId,
					Event.TargetClientIds: targetClientIds,
					Event.GroupId:         groupId,
				}),
			))
			return errors.New("websocketConnection does not exist")
		}
		if !server.websocketConnectionGroups[websocketId][groupId] {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientNotInGroup,
				"websocketConnection is not in group",
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
					Event.Behaviour:       Event.Transactional,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        websocketId,
					Event.TargetClientIds: targetClientIds,
					Event.GroupId:         groupId,
				}),
			))
			return errors.New("websocketConnection is not in group")
		}
	}

	for _, websocketId := range websocketIds {
		delete(server.websocketConnectionGroups[websocketId], groupId)
		delete(server.groupsWebsocketConnections[groupId], websocketId)
	}
	if len(server.groupsWebsocketConnections[groupId]) == 0 {
		delete(server.groupsWebsocketConnections, groupId)
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.AddedClientsToGroup,
		"removed websocketConnections from group",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
			Event.Behaviour:       Event.Transactional,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	))
	return nil
}

func (server *WebsocketServer) RemoveWebsocketConnectionsFromGroup_bestEffort(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	targetClientIds := Helpers.JsonMarshal(websocketIds)
	if event := server.onEvent(Event.NewInfo(
		Event.RemovingClientsFromGroup,
		"removing websocketConnections from group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
			Event.Behaviour:       Event.BestEffort,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
				Event.Behaviour:       Event.BestEffort,
				Event.ClientType:      Event.WebsocketConnection,
				Event.TargetClientIds: targetClientIds,
				Event.GroupId:         groupId,
			}),
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] != nil {
			delete(server.websocketConnectionGroups[websocketId], groupId)
		} else {
			event := server.onEvent(Event.NewWarning(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
					Event.Behaviour:       Event.BestEffort,
					Event.ClientType:      Event.WebsocketConnection,
					Event.ClientId:        websocketId,
					Event.TargetClientIds: targetClientIds,
					Event.GroupId:         groupId,
				}),
			))
			if !event.IsInfo() {
				return event.GetError()
			}
		}
		delete(server.groupsWebsocketConnections[groupId], websocketId)
	}

	if len(server.groupsWebsocketConnections[groupId]) == 0 {
		delete(server.groupsWebsocketConnections, groupId)
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.AddedClientsToGroup,
		"removed websocketConnections from group",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance:    Event.RemoveClientsFromGroupRoutine,
			Event.Behaviour:       Event.BestEffort,
			Event.ClientType:      Event.WebsocketConnection,
			Event.TargetClientIds: targetClientIds,
			Event.GroupId:         groupId,
		}),
	))
	return nil
}

func (server *WebsocketServer) GetGroupWebsocketConnectionIds(groupId string) ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingGroupClients,
		"getting group websocketConnections",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetGroupClientsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.GroupId:      groupId,
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance: Event.GetGroupClientsRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.GroupId:      groupId,
			}),
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
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetGroupClientsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.GroupId:      groupId,
			Event.Result:       Helpers.JsonMarshal(groupClientIds),
		}),
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
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetClientGroupsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	if server.websocketConnections[websocketId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance: Event.GetClientGroupsRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.ClientId:     websocketId,
			}),
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
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetClientGroupsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.Result:       Helpers.JsonMarshal(groupIds),
		}),
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
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetGroupCountRoutine,
			Event.ClientType:   Event.WebsocketConnection,
		}),
	)); !event.IsInfo() {
		return -1, event.GetError()
	}

	groupCount := len(server.groupsWebsocketConnections)

	if event := server.onEvent(Event.NewInfoNoOption(
		Event.GotGroupCount,
		"got group count",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetGroupCountRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.Result:       Helpers.JsonMarshal(groupCount),
		}),
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
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetGroupIdsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
		}),
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
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetGroupIdsRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.Result:       Helpers.JsonMarshal(groups),
		}),
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
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.IsClientInGroupRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.GroupId:      groupId,
		}),
	)); !event.IsInfo() {
		return false, event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance: Event.IsClientInGroupRoutine,
				Event.ClientType:   Event.WebsocketConnection,
				Event.ClientId:     websocketId,
				Event.GroupId:      groupId,
				Event.Result:       Helpers.JsonMarshal(false),
			}),
		))
		return false, errors.New("group does not exist")
	}

	if event := server.onEvent(Event.NewInfoNoOption(
		Event.GotIsClientInGroup,
		"got is websocketConnection in group",
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.IsClientInGroupRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.GroupId:      groupId,
			Event.Result:       Helpers.JsonMarshal(true),
		}),
	)); !event.IsInfo() {
		return false, event.GetError()
	}
	return true, nil
}
