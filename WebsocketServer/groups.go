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

	if event := server.onInfo(Event.NewInfo(
		Event.AddingClientsToGroup,
		"adding clients to group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"client does not exist",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					"groupId":               groupId,
				}),
			))
			return errors.New("client does not exist")
		}
		if server.clientGroups[websocketId][groupId] {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientAlreadyInGroup,
				"client is already in group",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					"groupId":               groupId,
				}),
			))
			return errors.New("client is already in group")
		}
	}

	if server.groups[groupId] == nil {
		if event := server.onWarning(Event.NewWarning(
			Event.CreatingGroup,
			"group does not exist",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:           Event.WebsocketConnection,
				"groupId":            groupId,
				"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			}),
		)); !event.IsInfo() {
			return event.GetError()
		}
		server.groups[groupId] = make(map[string]*WebsocketClient)
	}

	for _, websocketId := range websocketIds {
		server.groups[groupId][websocketId] = server.clients[websocketId]
		server.clientGroups[websocketId][groupId] = true
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientsAddedToGroup,
		"added clients to group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
		}),
	))
	return nil
}

// AttemptToAddClientsToGroup adds websocket clients to a group.
// Proceeds even if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AttemptToAddClientsToGroup(groupId string, websocketIds ...string) error {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.AddingClientsToGroup,
		"adding clients to group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		if event := server.onWarning(Event.NewWarning(
			Event.CreatingGroup,
			"group does not exist",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:           Event.WebsocketConnection,
				"groupId":            groupId,
				"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			}),
		)); !event.IsInfo() {
			return event.GetError()
		}
		server.groups[groupId] = make(map[string]*WebsocketClient)
	}

	for _, websocketId := range websocketIds {
		if server.clients[websocketId] != nil {
			server.groups[groupId][websocketId] = server.clients[websocketId]
			server.clientGroups[websocketId][groupId] = true
		} else {
			event := server.onWarning(Event.NewWarning(
				Event.ClientDoesNotExist,
				"client does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					"groupId":               groupId,
				}),
			))
			if !event.IsInfo() {
				return event.GetError()
			}
		}
	}

	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientsAddedToGroup,
		"added clients to group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
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

	if event := server.onInfo(Event.NewInfo(
		Event.RemovingClientsFromGroup,
		"removing clients from group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:  Event.WebsocketConnection,
				"groupId":   groupId,
				"websocket": Helpers.JsonMarshal(websocketIds),
			}),
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"client does not exist",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					"groupId":               groupId,
				}),
			))
			return errors.New("client does not exist")
		}
		if !server.clientGroups[websocketId][groupId] {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientNotInGroup,
				"client is not in group",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					"groupId":               groupId,
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

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientsAddedToGroup,
		"removed clients from group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
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

	if event := server.onInfo(Event.NewInfo(
		Event.RemovingClientsFromGroup,
		"removing clients from group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:  Event.WebsocketConnection,
				"groupId":   groupId,
				"websocket": Helpers.JsonMarshal(websocketIds),
			}),
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.clients[websocketId] != nil {
			delete(server.clientGroups[websocketId], groupId)
		} else {
			event := server.onWarning(Event.NewWarning(
				Event.ClientDoesNotExist,
				"client does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					"groupId":               groupId,
				}),
			))
			if !event.IsInfo() {
				return event.GetError()
			}
		}
		delete(server.groups[groupId], websocketId)
	}

	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientsAddedToGroup,
		"removed clients from group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:           Event.WebsocketConnection,
			"targetWebsocketIds": Helpers.JsonMarshal(websocketIds),
			"groupId":            groupId,
		}),
	))
	return nil
}

func (server *WebsocketServer) GetGroupClients(groupId string) ([]string, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingGroupClients,
		"getting group clients",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
			"groupId":  groupId,
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind: Event.WebsocketConnection,
				"groupId":  groupId,
			}),
		))
		return nil, errors.New("group does not exist")
	}

	groupMembers := make([]string, 0)
	for websocketId := range server.groups[groupId] {
		groupMembers = append(groupMembers, websocketId)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.GotGroupClients,
		"got group clients",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
			"groupId":  groupId,
		}),
	))
	return groupMembers, nil
}

func (server *WebsocketServer) GetClientGroups(websocketId string) ([]string, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingClientGroups,
		"getting client groups",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.TargetWebsocketId: websocketId,
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	if server.clients[websocketId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"client does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.TargetWebsocketId: websocketId,
			}),
		))
		return nil, errors.New("client does not exist")
	}

	if server.clientGroups[websocketId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientNotInGroup,
			"client is not in any group",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.TargetWebsocketId: websocketId,
			}),
		))
		return nil, errors.New("client is not in any group")
	}

	groups := make([]string, 0)
	for groupId := range server.clientGroups[websocketId] {
		groups = append(groups, groupId)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.GotClientGroups,
		"got client groups",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.TargetWebsocketId: websocketId,
		}),
	))
	return groups, nil
}

// GetGroupCount returns the number of groups.
func (server *WebsocketServer) GetGroupCount() (int, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingGroupCount,
		"getting group count",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
		}),
	)); !event.IsInfo() {
		return -1, event.GetError()
	}

	return len(server.groups), nil
}

func (server *WebsocketServer) GetGroupIds() ([]string, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingGroupIds,
		"getting group ids",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	groups := make([]string, 0)
	for groupId := range server.groups {
		groups = append(groups, groupId)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.GotGroupIds,
		"got group ids",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
		}),
	))
	return groups, nil
}

func (server *WebsocketServer) IsClientInGroup(groupId string, websocketId string) (bool, error) {
	server.clientMutex.RLock()
	defer server.clientMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingGroupClients,
		"getting group clients",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
			"groupId":  groupId,
		}),
	)); !event.IsInfo() {
		return false, event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind: Event.WebsocketConnection,
				"groupId":  groupId,
			}),
		))
		return false, errors.New("group does not exist")
	}

	if server.groups[groupId][websocketId] == nil {
		server.onInfo(Event.NewInfoNoOption(
			Event.ClientNotInGroup,
			"client is not in group",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.TargetWebsocketId: websocketId,
				"groupId":               groupId,
			}),
		))
		return false, errors.New("client is not in group")
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.GotGroupClients,
		"got group clients",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind: Event.WebsocketConnection,
			"groupId":  groupId,
		}),
	))
	return true, nil
}
