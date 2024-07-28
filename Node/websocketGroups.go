package Node

import "github.com/neutralusername/Systemge/Error"

func (node *Node) AddToWebsocketGroup(groupId string, websocketIds ...string) error {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.Lock()
		defer node.websocket.mutex.Unlock()
		for _, websocketId := range websocketIds {
			if node.websocket.clients[websocketId] == nil {
				return Error.New("websocketClient with id "+websocketId+" does not exist", nil)
			}
			if node.websocket.clientGroups[websocketId][groupId] {
				return Error.New("websocketClient with id "+websocketId+" is already in group "+groupId, nil)
			}
		}
		if node.websocket.groups[groupId] == nil {
			node.websocket.groups[groupId] = make(map[string]*WebsocketClient)
		}
		for _, websocketId := range websocketIds {
			node.websocket.groups[groupId][websocketId] = node.websocket.clients[websocketId]
			node.websocket.clientGroups[websocketId][groupId] = true
		}
		return nil
	}
	return Error.New("websocket component is not initialized", nil)
}

func (node *Node) RemoveFromWebsocketGroup(groupId string, websocketIds ...string) error {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.Lock()
		defer node.websocket.mutex.Unlock()
		if node.websocket.groups[groupId] == nil {
			return Error.New("group with id "+groupId+" does not exist", nil)
		}
		for _, websocketId := range websocketIds {
			if node.websocket.clients[websocketId] == nil {
				return Error.New("websocketClient with id "+websocketId+" does not exist", nil)
			}
			if !node.websocket.clientGroups[websocketId][groupId] {
				return Error.New("websocketClient with id "+websocketId+" is not in group "+groupId, nil)
			}
		}
		for _, websocketId := range websocketIds {
			delete(node.websocket.clientGroups[websocketId], groupId)
			delete(node.websocket.groups[groupId], websocketId)
		}
		if len(node.websocket.groups[groupId]) == 0 {
			delete(node.websocket.groups, groupId)
		}
		return nil
	}
	return Error.New("websocket component is not initialized", nil)
}

func (node *Node) GetWebsocketGroupClients(groupId string) []string {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.Lock()
		defer node.websocket.mutex.Unlock()
		if node.websocket.groups[groupId] == nil {
			return nil
		}
		groupMembers := make([]string, 0)
		for websocketId := range node.websocket.groups[groupId] {
			groupMembers = append(groupMembers, websocketId)
		}
		return groupMembers
	}
	return nil
}

func (node *Node) GetWebsocketClientGroups(websocketId string) []string {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.Lock()
		defer node.websocket.mutex.Unlock()
		if node.websocket.clients[websocketId] == nil {
			return nil
		}
		if node.websocket.clientGroups[websocketId] == nil {
			return nil
		}
		groups := make([]string, 0)
		for groupId := range node.websocket.clientGroups[websocketId] {
			groups = append(groups, groupId)
		}
		return groups
	}
	return nil
}

func (node *Node) IsInWebsocketGroup(groupId string, websocketId string) bool {
	if websocket := node.websocket; websocket != nil {
		node.websocket.mutex.Lock()
		defer node.websocket.mutex.Unlock()
		if node.websocket.groups[groupId] == nil {
			return false
		}
		if node.websocket.groups[groupId][websocketId] == nil {
			return false
		}
		return true
	}
	return false
}
