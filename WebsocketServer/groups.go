package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

// AddClientsToGroup adds websocket clients to a group.
// Returns an error if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AddClientsToGroup(groupId string, websocketIds ...string) *Event.Event {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	if event := server.onInfo(Event.New(
		Event.AddingClientsToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "adding clients to group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AddClientsToGroup",
		}),
	)); event.IsError() {
		return event
	}
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			return server.onError(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client does not exist",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "AddClientsToGroup",
				}),
			))
		}
		if server.clientGroups[websocketId][groupId] {
			return server.onError(Event.New(
				Event.ClientAlreadyInGroup,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client is already in group",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "AddClientsToGroup",
				}),
			))
		}
	}
	if server.groups[groupId] == nil {
		server.groups[groupId] = make(map[string]*WebsocketClient)
	}
	for _, websocketId := range websocketIds {
		server.groups[groupId][websocketId] = server.clients[websocketId]
		server.clientGroups[websocketId][groupId] = true
	}
	return server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "added clients to group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AddClientsToGroup",
		}),
	))
}

// AttemptToAddClientsToGroup adds websocket clients to a group.
// Proceeds even if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AttemptToAddClientsToGroup(groupId string, websocketIds ...string) *Event.Event {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	if event := server.onInfo(Event.New(
		Event.AddingClientsToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "adding clients to group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AttemptToAddClientsToGroup",
		}),
	)); event.IsError() {
		return event
	}
	if server.groups[groupId] == nil {
		server.groups[groupId] = make(map[string]*WebsocketClient)
	}
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] != nil {
			server.groups[groupId][websocketId] = server.clients[websocketId]
			server.clientGroups[websocketId][groupId] = true
		} else {
			event := server.onWarning(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"warning":           "client does not exist",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "AttemptToAddClientsToGroup",
				}),
			))
			if event.IsError() {
				return event
			}
		}
	}
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}
	return server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "added clients to group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AttemptToAddClientsToGroup",
		}),
	))
}

// RemoveClientsFromGroup removes websocket clients from a group.
// Returns an error if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) RemoveClientsFromGroup(groupId string, websocketIds ...string) *Event.Event {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	if event := server.onInfo(Event.New(
		Event.RemovingClientsFromGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "removing clients from group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "RemoveClientsFromGroup",
		}),
	)); event.IsError() {
		return event
	}
	if server.groups[groupId] == nil {
		return server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":     "group does not exist",
				"type":      "websocketConnection",
				"groupId":   groupId,
				"websocket": Helpers.JsonMarshal(websocketIds),
				"function":  "RemoveClientsFromGroup",
			}),
		))
	}
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			return server.onError(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client does not exist",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "RemoveClientsFromGroup",
				}),
			))
		}
		if !server.clientGroups[websocketId][groupId] {
			return server.onError(Event.New(
				Event.ClientNotInGroup,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client is not in group",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "RemoveClientsFromGroup",
				}),
			))
		}
	}
	for _, websocketId := range websocketIds {
		delete(server.clientGroups[websocketId], groupId)
		delete(server.groups[groupId], websocketId)
	}
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}
	return server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "removed clients from group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "RemoveClientsFromGroup",
		}),
	))
}

// AttemptToRemoveClientsFromGroup removes websocket clients from a group.
// proceeds even if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) AttemptToRemoveClientsFromGroup(groupId string, websocketIds ...string) *Event.Event {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	if event := server.onInfo(Event.New(
		Event.RemovingClientsFromGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "removing clients from group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AttemptToRemoveClientsFromGroup",
		}),
	)); event.IsError() {
		return event
	}
	if server.groups[groupId] == nil {
		return server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":     "group does not exist",
				"type":      "websocketConnection",
				"groupId":   groupId,
				"websocket": Helpers.JsonMarshal(websocketIds),
				"function":  "AttemptToRemoveClientsFromGroup",
			}),
		))
	}
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] != nil {
			delete(server.clientGroups[websocketId], groupId)
		} else {
			event := server.onWarning(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"warning":           "client does not exist",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "AttemptToRemoveClientsFromGroup",
				}),
			))
			if event.IsError() {
				return event
			}
		}
		delete(server.groups[groupId], websocketId)
	}
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}
	return server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "removed clients from group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AttemptToRemoveClientsFromGroup",
		}),
	))
}

func (server *WebsocketServer) GetGroupClients(groupId string) []string {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	if server.groups[groupId] == nil {
		return nil
	}
	groupMembers := make([]string, 0)
	for websocketId := range server.groups[groupId] {
		groupMembers = append(groupMembers, websocketId)
	}
	return groupMembers
}

func (server *WebsocketServer) GetClientGroups(websocketId string) []string {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	if server.clients[websocketId] == nil {
		return nil
	}
	if server.clientGroups[websocketId] == nil {
		return nil
	}
	groups := make([]string, 0)
	for groupId := range server.clientGroups[websocketId] {
		groups = append(groups, groupId)
	}
	return groups
}

// GetGroupCount returns the number of groups.
func (server *WebsocketServer) GetGroupCount() int {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	return len(server.groups)
}

func (server *WebsocketServer) GetGroups() []string {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	groups := make([]string, 0)
	for groupId := range server.groups {
		groups = append(groups, groupId)
	}
	return groups
}

func (server *WebsocketServer) IsClientInGroup(groupId string, websocketId string) bool {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()
	if server.groups[groupId] == nil {
		return false
	}
	if server.groups[groupId][websocketId] == nil {
		return false
	}
	return true
}
