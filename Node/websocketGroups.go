package Node

import "Systemge/Error"

func (node *Node) AddToWebsocketGroup(groupId string, websocketId string) error {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if node.websocketClients[websocketId] == nil {
		return Error.New("websocketClient with id "+websocketId+" does not exist", nil)
	}
	if node.websocketClientGroups[websocketId][groupId] {
		return Error.New("websocketClient with id "+websocketId+" is already in group "+groupId, nil)
	}
	if node.websocketGroups[groupId] == nil {
		node.websocketGroups[groupId] = make(map[string]*WebsocketClient)
	}
	node.websocketGroups[groupId][websocketId] = node.websocketClients[websocketId]
	node.websocketClientGroups[websocketId][groupId] = true
	return nil
}

func (node *Node) RemoveFromWebsocketGroup(groupId string, websocketId string) error {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if node.websocketGroups[groupId] == nil {
		return Error.New("group with id "+groupId+" does not exist", nil)
	}
	if node.websocketGroups[groupId][websocketId] == nil {
		return Error.New("websocketClient with id "+websocketId+" is not in group "+groupId, nil)
	}
	if node.websocketClients[websocketId] == nil {
		return Error.New("websocketClient with id "+websocketId+" does not exist", nil)
	}
	if node.websocketClientGroups[websocketId] == nil {
		return Error.New("websocketClient with id "+websocketId+" is not in any groups", nil)
	}
	delete(node.websocketClientGroups[websocketId], groupId)
	delete(node.websocketGroups[groupId], websocketId)
	if len(node.websocketGroups[groupId]) == 0 {
		delete(node.websocketGroups, groupId)
	}
	return nil
}

func (node *Node) GetWebsocketGroupClients(groupId string) []string {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if node.websocketGroups[groupId] == nil {
		return nil
	}
	groupMembers := make([]string, 0)
	for websocketId := range node.websocketGroups[groupId] {
		groupMembers = append(groupMembers, websocketId)
	}
	return groupMembers
}

func (node *Node) GetWebsocketGroups(websocketId string) []string {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if node.websocketClients[websocketId] == nil {
		return nil
	}
	if node.websocketClientGroups[websocketId] == nil {
		return nil
	}
	groups := make([]string, 0)
	for groupId := range node.websocketClientGroups[websocketId] {
		groups = append(groups, groupId)
	}
	return groups
}

func (node *Node) IsInWebsocketGroup(groupId string, websocketId string) bool {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if node.websocketGroups[groupId] == nil {
		return false
	}
	if node.websocketGroups[groupId][websocketId] == nil {
		return false
	}
	return true
}
