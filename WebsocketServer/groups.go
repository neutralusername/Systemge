package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

// AddClientsToGroup adds websocket websocketConnections to a group.
// Returns an error if either of the websocket websocketConnections does not exist or is already in the group.
func (server *WebsocketServer) AddClientsToGroup(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.AddingClientsToGroup,
		"adding websocketConnection to group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] == nil {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					Event.GroupId:           groupId,
				}),
			))
			return errors.New("client does not exist")
		}
		if server.websocketConnectionGroups[websocketId][groupId] {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientAlreadyInGroup,
				"websocketConnection is already in group",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					Event.GroupId:           groupId,
				}),
			))
			return errors.New("websocketConnection is already in group")
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
				Event.Kind:               Event.WebsocketConnection,
				Event.GroupId:            groupId,
				Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			}),
		)); !event.IsInfo() {
			return event.GetError()
		}
		server.groups[groupId] = make(map[string]*WebsocketConnection)
	}

	for _, websocketId := range websocketIds {
		server.groups[groupId][websocketId] = server.websocketConnections[websocketId]
		server.websocketConnectionGroups[websocketId][groupId] = true
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientsAddedToGroup,
		"added websocketConnections to group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
		}),
	))
	return nil
}

// AttemptToAddClientsToGroup adds websocket websocketConnections to a group.
// Proceeds even if either of the websocket websocketConnections does not exist or is already in the group.
func (server *WebsocketServer) AttemptToAddClientsToGroup(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.AddingClientsToGroup,
		"adding websocketConnections to group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
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
				Event.Kind:               Event.WebsocketConnection,
				Event.GroupId:            groupId,
				Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			}),
		)); !event.IsInfo() {
			return event.GetError()
		}
		server.groups[groupId] = make(map[string]*WebsocketConnection)
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] != nil {
			server.groups[groupId][websocketId] = server.websocketConnections[websocketId]
			server.websocketConnectionGroups[websocketId][groupId] = true
		} else {
			event := server.onWarning(Event.NewWarning(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					Event.GroupId:           groupId,
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
		"added websocketConnections to group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
		}),
	))
	return nil
}

// RemoveClientsFromGroup removes websocket websocketConnections from a group.
// Returns an error if either of the websocket websocketConnections does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) RemoveClientsFromGroup(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.RemovingClientsFromGroup,
		"removing websocketConnections from group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:               Event.WebsocketConnection,
				Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
				Event.GroupId:            groupId,
			}),
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] == nil {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					Event.GroupId:           groupId,
				}),
			))
			return errors.New("websocketConnection does not exist")
		}
		if !server.websocketConnectionGroups[websocketId][groupId] {
			server.onWarning(Event.NewWarningNoOption(
				Event.ClientNotInGroup,
				"websocketConnection is not in group",
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					Event.GroupId:           groupId,
				}),
			))
			return errors.New("websocketConnection is not in group")
		}
	}

	for _, websocketId := range websocketIds {
		delete(server.websocketConnectionGroups[websocketId], groupId)
		delete(server.groups[groupId], websocketId)
	}
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.ClientsAddedToGroup,
		"removed websocketConnections from group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
		}),
	))
	return nil
}

// AttemptToRemoveClientsFromGroup removes websocket websocketConnections from a group.
// proceeds even if either of the websocket websocketConnections does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) AttemptToRemoveClientsFromGroup(groupId string, websocketIds ...string) error {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()

	if event := server.onInfo(Event.NewInfo(
		Event.RemovingClientsFromGroup,
		"removing websocketConnections from group",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:               Event.WebsocketConnection,
				Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
				Event.GroupId:            groupId,
			}),
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] != nil {
			delete(server.websocketConnectionGroups[websocketId], groupId)
		} else {
			event := server.onWarning(Event.NewWarning(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				server.GetServerContext().Merge(Event.Context{
					Event.Kind:              Event.WebsocketConnection,
					Event.TargetWebsocketId: websocketId,
					Event.GroupId:           groupId,
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
		"removed websocketConnections from group",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:               Event.WebsocketConnection,
			Event.TargetWebsocketIds: Helpers.JsonMarshal(websocketIds),
			Event.GroupId:            groupId,
		}),
	))
	return nil
}

func (server *WebsocketServer) GetGroupClients(groupId string) ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingGroupClients,
		"getting group websocketConnections",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    Event.WebsocketConnection,
			Event.GroupId: groupId,
		}),
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:    Event.WebsocketConnection,
				Event.GroupId: groupId,
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
		"got group websocketConnections",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    Event.WebsocketConnection,
			Event.GroupId: groupId,
		}),
	))
	return groupMembers, nil
}

func (server *WebsocketServer) GetClientGroups(websocketId string) ([]string, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingClientGroups,
		"getting websocketConnection groups",
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

	if server.websocketConnections[websocketId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"websocketConnection does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.TargetWebsocketId: websocketId,
			}),
		))
		return nil, errors.New("websocketConnection does not exist")
	}

	if server.websocketConnectionGroups[websocketId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.ClientNotInGroup,
			"websocketConnection is not in any group",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.TargetWebsocketId: websocketId,
			}),
		))
		return nil, errors.New("websocketConnection is not in any group")
	}

	groups := make([]string, 0)
	for groupId := range server.websocketConnectionGroups[websocketId] {
		groups = append(groups, groupId)
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.GotClientGroups,
		"got websocketConnection groups",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:              Event.WebsocketConnection,
			Event.TargetWebsocketId: websocketId,
		}),
	))
	return groups, nil
}

// GetGroupCount returns the number of groups.
func (server *WebsocketServer) GetGroupCount() (int, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

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
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

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
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onInfo(Event.NewInfo(
		Event.GettingGroupClients,
		"getting group websocketConnections",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    Event.WebsocketConnection,
			Event.GroupId: groupId,
		}),
	)); !event.IsInfo() {
		return false, event.GetError()
	}

	if server.groups[groupId] == nil {
		server.onWarning(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:    Event.WebsocketConnection,
				Event.GroupId: groupId,
			}),
		))
		return false, errors.New("group does not exist")
	}

	if server.groups[groupId][websocketId] == nil {
		server.onInfo(Event.NewInfoNoOption(
			Event.ClientNotInGroup,
			"websocketConnection is not in group",
			server.GetServerContext().Merge(Event.Context{
				Event.Kind:              Event.WebsocketConnection,
				Event.TargetWebsocketId: websocketId,
				Event.GroupId:           groupId,
			}),
		))
		return false, errors.New("client is not in group")
	}

	server.onInfo(Event.NewInfoNoOption(
		Event.GotGroupClients,
		"got group websocketConnections",
		server.GetServerContext().Merge(Event.Context{
			Event.Kind:    Event.WebsocketConnection,
			Event.GroupId: groupId,
		}),
	))
	return true, nil
}
