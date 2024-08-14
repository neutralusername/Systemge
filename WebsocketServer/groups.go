package WebsocketServer

import "github.com/neutralusername/Systemge/Error"

// AddClientToGroup adds websocket clients to a group.
// Returns an error if either of the websocket clients does not exist or is already in the group.
func (server *WebsocketServer) AddClientToGroup(groupId string, websocketIds ...string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, websocketId := range websocketIds {
		if server.clients[websocketId] == nil {
			return Error.New("client with id "+websocketId+" does not exist", nil)
		}
		if server.clientGroups[websocketId][groupId] {
			return Error.New("client with id "+websocketId+" is already in group "+groupId, nil)
		}
	}
	if server.groups[groupId] == nil {
		server.groups[groupId] = make(map[string]*Client)
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
	if websocket := server; websocket != nil {
		server.mutex.Lock()
		defer server.mutex.Unlock()
		if server.groups[groupId] == nil {
			server.groups[groupId] = make(map[string]*Client)
		}
		for _, websocketId := range websocketIds {
			if server.clients[websocketId] != nil {
				server.groups[groupId][websocketId] = server.clients[websocketId]
				server.clientGroups[websocketId][groupId] = true
			}
		}
		return nil
	}
	return Error.New("websocket component is not initialized", nil)
}

// RemoveClientFromGroup_ removes websocket clients from a group.
// proceeds even if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) RemoveClientFromGroup_(groupId string, websocketIds ...string) error {
	if websocket := server; websocket != nil {
		server.mutex.Lock()
		defer server.mutex.Unlock()
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
	return Error.New("websocket component is not initialized", nil)
}

// RemoveClientFromGroup removes websocket clients from a group.
// Returns an error if either of the websocket clients does not exist or is not in the group.
// Returns an error if the group does not exist.
func (server *WebsocketServer) RemoveClientFromGroup(groupId string, websocketIds ...string) error {
	if websocket := server; websocket != nil {
		server.mutex.Lock()
		defer server.mutex.Unlock()
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
	return Error.New("websocket component is not initialized", nil)
}

func (server *WebsocketServer) GetGroupClients(groupId string) []string {
	if websocket := server; websocket != nil {
		server.mutex.RLock()
		defer server.mutex.RUnlock()
		if server.groups[groupId] == nil {
			return nil
		}
		groupMembers := make([]string, 0)
		for websocketId := range server.groups[groupId] {
			groupMembers = append(groupMembers, websocketId)
		}
		return groupMembers
	}
	return nil
}

func (server *WebsocketServer) GetClientGroups(websocketId string) []string {
	server.mutex.RLock()
	defer server.mutex.RUnlock()
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
	server.mutex.RLock()
	defer server.mutex.RUnlock()
	if server.groups[groupId] == nil {
		return false
	}
	if server.groups[groupId][websocketId] == nil {
		return false
	}
	return true
}
