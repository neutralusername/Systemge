package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

// AddClientsToGroup adds websocket clients to a group.
// Returns an error if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AddClientsToGroup(groupId string, websocketIds ...string) error {
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
			"onError":            "cancel",
			"onWarning":          "continue",
			"onInfo":             "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}

	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			server.onError(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client does not exist",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "AddClientsToGroup",
				}),
			))
			return errors.New("client does not exist")
		}
		if server.clientGroups[websocketId][groupId] {
			server.onError(Event.New(
				Event.ClientAlreadyInGroup,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client is already in group",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "AddClientsToGroup",
				}),
			))
			return errors.New("client is already in group")
		}
	}

	if server.groups[groupId] == nil {
		if event := server.onWarning(Event.New(
			Event.CreatingGroup,
			server.GetServerContext().Merge(Event.Context{
				"warning":            "group does not exist",
				"type":               "websocketConnection",
				"groupId":            groupId,
				"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
				"function":           "AddClientsToGroup",
				"onError":            "cancel",
				"onWarning":          "continue",
				"onInfo":             "continue",
			}),
		)); event.IsError() {
			return event.GetError()
		}
		server.groups[groupId] = make(map[string]*WebsocketClient)
	}

	for _, websocketId := range websocketIds {
		server.groups[groupId][websocketId] = server.clients[websocketId]
		server.clientGroups[websocketId][groupId] = true
	}

	server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "added clients to group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AddClientsToGroup",
		}),
	))
	return nil
}

// AttemptToAddClientsToGroup adds websocket clients to a group.
// Proceeds even if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AttemptToAddClientsToGroup(groupId string, websocketIds ...string) error {
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
			"onError":            "cancel",
			"onWarning":          "continue",
			"onInfo":             "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		if event := server.onWarning(Event.New(
			Event.CreatingGroup,
			server.GetServerContext().Merge(Event.Context{
				"warning":            "group does not exist",
				"type":               "websocketConnection",
				"groupId":            groupId,
				"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
				"function":           "AttemptToAddClientsToGroup",
			}),
		)); event.IsError() {
			return event.GetError()
		}
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
					"onError":           "cancel",
					"onWarning":         "continue",
					"onInfo":            "continue",
				}),
			))
			if event.IsError() {
				return event.GetError()
			}
		}
	}

	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}

	server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "added clients to group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AttemptToAddClientsToGroup",
		}),
	))
	return nil
}

// RemoveClientsFromGroup removes websocket clients from a group.
// Returns an error if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) RemoveClientsFromGroup(groupId string, websocketIds ...string) error {
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
			"onError":            "cancel",
			"onWarning":          "continue",
			"onInfo":             "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":     "group does not exist",
				"type":      "websocketConnection",
				"groupId":   groupId,
				"websocket": Helpers.JsonMarshal(websocketIds),
				"function":  "RemoveClientsFromGroup",
			}),
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			server.onError(Event.New(
				Event.ClientDoesNotExist,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client does not exist",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "RemoveClientsFromGroup",
				}),
			))
			return errors.New("client does not exist")
		}
		if !server.clientGroups[websocketId][groupId] {
			server.onError(Event.New(
				Event.ClientNotInGroup,
				server.GetServerContext().Merge(Event.Context{
					"error":             "client is not in group",
					"type":              "websocketConnection",
					"targetWebsocketId": websocketId,
					"groupId":           groupId,
					"function":          "RemoveClientsFromGroup",
				}),
			))
			return errors.New("client is not in group")
		}
	}

	for _, websocketId := range websocketIds {
		delete(server.clientGroups[websocketId], groupId)
		delete(server.groups[groupId], websocketId)
	}
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}

	server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "removed clients from group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "RemoveClientsFromGroup",
		}),
	))
	return nil
}

// AttemptToRemoveClientsFromGroup removes websocket clients from a group.
// proceeds even if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) AttemptToRemoveClientsFromGroup(groupId string, websocketIds ...string) error {
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
			"onError":            "cancel",
			"onWarning":          "continue",
			"onInfo":             "continue",
		}),
	)); event.IsError() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":     "group does not exist",
				"type":      "websocketConnection",
				"groupId":   groupId,
				"websocket": Helpers.JsonMarshal(websocketIds),
				"function":  "AttemptToRemoveClientsFromGroup",
			}),
		))
		return errors.New("group does not exist")
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
					"onError":           "cancel",
					"onWarning":         "continue",
					"onInfo":            "continue",
				}),
			))
			if event.IsError() {
				return event.GetError()
			}
		}
		delete(server.groups[groupId], websocketId)
	}

	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}

	server.onInfo(Event.New(
		Event.ClientsAddedToGroup,
		server.GetServerContext().Merge(Event.Context{
			"info":               "removed clients from group",
			"type":               "websocketConnection",
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
			"function":           "AttemptToRemoveClientsFromGroup",
		}),
	))
	return nil
}

func (server *WebsocketServer) GetGroupClients(groupId string) ([]string, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.New(
		Event.GettingGroupClients,
		server.GetServerContext().Merge(Event.Context{
			"info":      "getting group clients",
			"type":      "websocketConnection",
			"groupId":   groupId,
			"function":  "GetGroupClients",
			"onError":   "cancel",
			"onWarning": "continue",
			"onInfo":    "continue",
		}),
	)); event.IsError() {
		return nil, event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":    "group does not exist",
				"type":     "websocketConnection",
				"groupId":  groupId,
				"function": "GetGroupClients",
			}),
		))
		return nil, errors.New("group does not exist")
	}

	groupMembers := make([]string, 0)
	for websocketId := range server.groups[groupId] {
		groupMembers = append(groupMembers, websocketId)
	}

	server.onInfo(Event.New(
		Event.GotGroupClients,
		server.GetServerContext().Merge(Event.Context{
			"info":     "got group clients",
			"type":     "websocketConnection",
			"groupId":  groupId,
			"function": "GetGroupClients",
		}),
	))
	return groupMembers, nil
}

func (server *WebsocketServer) GetClientGroups(websocketId string) ([]string, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.New(
		Event.GettingClientGroups,
		server.GetServerContext().Merge(Event.Context{
			"info":              "getting client groups",
			"type":              "websocketConnection",
			"targetWebsocketId": websocketId,
			"function":          "GetClientGroups",
			"onError":           "cancel",
			"onWarning":         "continue",
			"onInfo":            "continue",
		}),
	)); event.IsError() {
		return nil, event.GetError()
	}

	if server.clients[websocketId] == nil {
		server.onError(Event.New(
			Event.ClientDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":             "client does not exist",
				"type":              "websocketConnection",
				"targetWebsocketId": websocketId,
				"function":          "GetClientGroups",
			}),
		))
		return nil, errors.New("client does not exist")
	}

	if server.clientGroups[websocketId] == nil {
		server.onError(Event.New(
			Event.ClientNotInGroup,
			server.GetServerContext().Merge(Event.Context{
				"error":             "client is not in any group",
				"type":              "websocketConnection",
				"targetWebsocketId": websocketId,
				"function":          "GetClientGroups",
			}),
		))
		return nil, errors.New("client is not in any group")
	}

	groups := make([]string, 0)
	for groupId := range server.clientGroups[websocketId] {
		groups = append(groups, groupId)
	}

	server.onInfo(Event.New(
		Event.GotClientGroups,
		server.GetServerContext().Merge(Event.Context{
			"info":              "got client groups",
			"type":              "websocketConnection",
			"targetWebsocketId": websocketId,
			"function":          "GetClientGroups",
		}),
	))
	return groups, nil
}

// GetGroupCount returns the number of groups.
func (server *WebsocketServer) GetGroupCount() (int, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.New(
		Event.GettingGroupCount,
		server.GetServerContext().Merge(Event.Context{
			"info":      "getting group count",
			"type":      "websocketConnection",
			"onError":   "cancel",
			"onWarning": "continue",
			"onInfo":    "continue",
		}),
	)); event.IsError() {
		return -1, event.GetError()
	}

	return len(server.groups), nil
}

func (server *WebsocketServer) GetGroupIds() ([]string, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.New(
		Event.GettingGroupIds,
		server.GetServerContext().Merge(Event.Context{
			"info":      "getting group ids",
			"type":      "websocketConnection",
			"onError":   "cancel",
			"onWarning": "continue",
			"onInfo":    "continue",
		}),
	)); event.IsError() {
		return nil, event.GetError()
	}

	groups := make([]string, 0)
	for groupId := range server.groups {
		groups = append(groups, groupId)
	}

	server.onInfo(Event.New(
		Event.GotGroupIds,
		server.GetServerContext().Merge(Event.Context{
			"info": "got group ids",
			"type": "websocketConnection",
		}),
	))
	return groups, nil
}

func (server *WebsocketServer) IsClientInGroup(groupId string, websocketId string) (bool, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.New(
		Event.GettingGroupClients,
		server.GetServerContext().Merge(Event.Context{
			"info":      "getting group clients",
			"type":      "websocketConnection",
			"groupId":   groupId,
			"function":  "IsClientInGroup",
			"onError":   "cancel",
			"onWarning": "continue",
			"onInfo":    "continue",
		}),
	)); event.IsError() {
		return false, event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onError(Event.New(
			Event.GroupDoesNotExist,
			server.GetServerContext().Merge(Event.Context{
				"error":    "group does not exist",
				"type":     "websocketConnection",
				"groupId":  groupId,
				"function": "IsClientInGroup",
			}),
		))
		return false, errors.New("group does not exist")
	}

	if server.groups[groupId][websocketId] == nil {
		server.onInfo(Event.New(
			Event.ClientNotInGroup,
			server.GetServerContext().Merge(Event.Context{
				"info":              "client is not in group",
				"type":              "websocketConnection",
				"targetWebsocketId": websocketId,
				"groupId":           groupId,
				"function":          "IsClientInGroup",
			}),
		))
		return false, errors.New("client is not in group")
	}

	server.onInfo(Event.New(
		Event.GotGroupClients,
		server.GetServerContext().Merge(Event.Context{
			"info":     "got group clients",
			"type":     "websocketConnection",
			"groupId":  groupId,
			"function": "IsClientInGroup",
		}),
	))
	return true, nil
}
