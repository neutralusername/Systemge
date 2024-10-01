package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

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
		Event.Context{
			Event.Circumstance:     Event.RemoveClientsFromGroup,
			Event.Behaviour:        Event.Transactional,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Context{
				Event.Circumstance:     Event.RemoveClientsFromGroup,
				Event.Behaviour:        Event.Transactional,
				Event.IdentityType:     Event.WebsocketConnection,
				Event.TargetIdentities: targetClientIds,
				Event.GroupId:          groupId,
			},
		))
		return errors.New("group does not exist")
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] == nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Context{
					Event.Circumstance:     Event.RemoveClientsFromGroup,
					Event.Behaviour:        Event.Transactional,
					Event.IdentityType:     Event.WebsocketConnection,
					Event.ClientId:         websocketId,
					Event.TargetIdentities: targetClientIds,
					Event.GroupId:          groupId,
				},
			))
			return errors.New("websocketConnection does not exist")
		}
		if !server.websocketConnectionGroups[websocketId][groupId] {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientNotInGroup,
				"websocketConnection is not in group",
				Event.Context{
					Event.Circumstance:     Event.RemoveClientsFromGroup,
					Event.Behaviour:        Event.Transactional,
					Event.IdentityType:     Event.WebsocketConnection,
					Event.ClientId:         websocketId,
					Event.TargetIdentities: targetClientIds,
					Event.GroupId:          groupId,
				},
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
		Event.Context{
			Event.Circumstance:     Event.RemoveClientsFromGroup,
			Event.Behaviour:        Event.Transactional,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
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
		Event.Context{
			Event.Circumstance:     Event.RemoveClientsFromGroup,
			Event.Behaviour:        Event.BestEffort,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.groupsWebsocketConnections[groupId] == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.GroupDoesNotExist,
			"group does not exist",
			Event.Context{
				Event.Circumstance:     Event.RemoveClientsFromGroup,
				Event.Behaviour:        Event.BestEffort,
				Event.IdentityType:     Event.WebsocketConnection,
				Event.TargetIdentities: targetClientIds,
				Event.GroupId:          groupId,
			},
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
				Event.Context{
					Event.Circumstance:     Event.RemoveClientsFromGroup,
					Event.Behaviour:        Event.BestEffort,
					Event.IdentityType:     Event.WebsocketConnection,
					Event.ClientId:         websocketId,
					Event.TargetIdentities: targetClientIds,
					Event.GroupId:          groupId,
				},
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
		Event.Context{
			Event.Circumstance:     Event.RemoveClientsFromGroup,
			Event.Behaviour:        Event.BestEffort,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
	))
	return nil
}
