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
		Event.Context{
			Event.Circumstance:     Event.AddClientsToGroup,
			Event.Behaviour:        Event.Transactional,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	for _, websocketId := range websocketIds {
		if server.websocketConnections[websocketId] == nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientDoesNotExist,
				"websocketConnection does not exist",
				Event.Context{
					Event.Circumstance:     Event.AddClientsToGroup,
					Event.Behaviour:        Event.Transactional,
					Event.IdentityType:     Event.WebsocketConnection,
					Event.ClientId:         websocketId,
					Event.TargetIdentities: targetClientIds,
					Event.GroupId:          groupId,
				},
			))
			return errors.New("client does not exist")
		}
		if server.websocketConnectionGroups[websocketId][groupId] {
			server.onEvent(Event.NewWarningNoOption(
				Event.ClientAlreadyInGroup,
				"websocketConnection is already in group",
				Event.Context{
					Event.Circumstance:     Event.AddClientsToGroup,
					Event.Behaviour:        Event.Transactional,
					Event.IdentityType:     Event.WebsocketConnection,
					Event.ClientId:         websocketId,
					Event.TargetIdentities: targetClientIds,
					Event.GroupId:          groupId,
				},
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
			Event.Context{
				Event.Circumstance:     Event.AddClientsToGroup,
				Event.Behaviour:        Event.Transactional,
				Event.IdentityType:     Event.WebsocketConnection,
				Event.TargetIdentities: targetClientIds,
				Event.GroupId:          groupId,
			},
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
		Event.Context{
			Event.Circumstance:     Event.AddClientsToGroup,
			Event.Behaviour:        Event.Transactional,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
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
		Event.Context{
			Event.Circumstance:     Event.AddClientsToGroup,
			Event.Behaviour:        Event.BestEffort,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
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
			Event.Context{
				Event.Circumstance:     Event.AddClientsToGroup,
				Event.Behaviour:        Event.BestEffort,
				Event.IdentityType:     Event.WebsocketConnection,
				Event.TargetIdentities: targetClientIds,
				Event.GroupId:          groupId,
			},
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
				Event.Context{
					Event.Circumstance:     Event.AddClientsToGroup,
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
	}

	if len(server.groupsWebsocketConnections[groupId]) == 0 {
		delete(server.groupsWebsocketConnections, groupId)
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.AddedClientsToGroup,
		"added websocketConnections to group",
		Event.Context{
			Event.Circumstance:     Event.AddClientsToGroup,
			Event.Behaviour:        Event.BestEffort,
			Event.IdentityType:     Event.WebsocketConnection,
			Event.TargetIdentities: targetClientIds,
			Event.GroupId:          groupId,
		},
	))
	return nil
}
