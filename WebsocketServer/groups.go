package WebsocketServer

import "github.com/neutralusername/Systemge/Error"

// AddClientToGroup adds websocket clients to a group.
// Returns an error if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AddClientToGroup(groupId string, websocketIds ...string) error {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			return Error.New("client with id "+websocketId+" does not exist", nil)
		}
		if server.clientGroups[websocketId][groupId] {
			return Error.New("client with id "+websocketId+" is already in group "+groupId, nil)
		}
	}
	if server.groups[groupId] == nil {
		server.groups[groupId] = make(map[string]*WebsocketClient)
	}
	for _, websocketId := range websocketIds {
		server.groups[groupId][websocketId] = server.clients[websocketId]
		server.clientGroups[websocketId][groupId] = true
	}
	return nil
}

// AddClientToGroup_ adds websocket clients to a group.
// Proceeds even if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AddClientToGroup_(groupId string, websocketIds ...string) error {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	if server.groups[groupId] == nil {
		server.groups[groupId] = make(map[string]*WebsocketClient)
	}
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] != nil {
			server.groups[groupId][websocketId] = server.clients[websocketId]
			server.clientGroups[websocketId][groupId] = true
		}
	}
	return nil
}

// RemoveClientFromGroup_ removes websocket clients from a group.
// proceeds even if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) RemoveClientFromGroup_(groupId string, websocketIds ...string) error {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	if server.groups[groupId] == nil {
		return Error.New("group with id "+groupId+" does not exist", nil)
	}
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] != nil {
			delete(server.clientGroups[websocketId], groupId)
		}
		if server.groups[groupId] != nil {
			delete(server.groups[groupId], websocketId)
		}
	}
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}
	return nil
}

// RemoveClientFromGroup removes websocket clients from a group.
// Returns an error if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) RemoveClientFromGroup(groupId string, websocketIds ...string) error {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	if server.groups[groupId] == nil {
		return Error.New("group with id "+groupId+" does not exist", nil)
	}
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			return Error.New("client with id "+websocketId+" does not exist", nil)
		}
		if !server.clientGroups[websocketId][groupId] {
			return Error.New("client with id "+websocketId+" is not in group "+groupId, nil)
		}
	}
	for _, websocketId := range websocketIds {
		delete(server.clientGroups[websocketId], groupId)
		delete(server.groups[groupId], websocketId)
	}
	if len(server.groups[groupId]) == 0 {
		delete(server.groups, groupId)
	}
	return nil
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

func (server *WebsocketServer) IsInWebsocketGroup(groupId string, websocketId string) bool {
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
