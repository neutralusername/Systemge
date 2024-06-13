package WebsocketServer

import (
	"Systemge/Error"
	"Systemge/WebsocketClient"
)

func (server *Server) AddToGroup(groupId string, websocketId string) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	client := server.clients[websocketId]
	if client == nil {
		return Error.New("WebsocketClient with id "+websocketId+" does not exist", nil)
	}
	if server.clientGroups[websocketId][groupId] {
		return Error.New("WebsocketClient with id "+websocketId+" is already in group "+groupId, nil)
	}
	if server.groups[groupId] == nil {
		server.groups[groupId] = make(map[string]*WebsocketClient.Client)
	}
	server.groups[groupId][websocketId] = server.clients[websocketId]
	server.clientGroups[websocketId][groupId] = true
	return nil
}

func (server *Server) RemoveFromGroup(groupId string, websocketId string) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if server.groups[groupId] == nil {
		return Error.New("Group with id "+groupId+" does not exist", nil)
	}
	if server.groups[groupId][websocketId] == nil {
		return Error.New("WebsocketClient with id "+websocketId+" is not in group "+groupId, nil)
	}
	if server.clients[websocketId] == nil {
		return Error.New("WebsocketClient with id "+websocketId+" does not exist", nil)
	}
	if server.clientGroups[websocketId] == nil {
		return Error.New("WebsocketClient with id "+websocketId+" is not in any groups", nil)
	}
	delete(server.clientGroups[websocketId], groupId)
	delete(server.groups[groupId], websocketId)
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}
	return nil
}

func (server *Server) GetGroupMembers(groupId string) []string {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if server.groups[groupId] == nil {
		return nil
	}
	groupMembers := make([]string, 0)
	for websocketId := range server.groups[groupId] {
		groupMembers = append(groupMembers, websocketId)
	}
	return groupMembers
}

func (server *Server) GetGroups(websocketId string) []string {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
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

func (server *Server) IsInGroup(groupId string, websocketId string) bool {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if server.groups[groupId] == nil {
		return false
	}
	if server.groups[groupId][websocketId] == nil {
		return false
	}
	return true
}
