package Client

import (
	"Systemge/Utilities"
	"Systemge/WebsocketClient"
)

func (client *Client) AddToGroup(groupId string, websocketId string) error {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	websocketClient := client.websocketClients[websocketId]
	if websocketClient == nil {
		return Utilities.NewError("WebsocketClient with id "+websocketId+" does not exist", nil)
	}
	if client.websocketClientGroups[websocketId][groupId] {
		return Utilities.NewError("WebsocketClient with id "+websocketId+" is already in group "+groupId, nil)
	}
	if client.groups[groupId] == nil {
		client.groups[groupId] = make(map[string]*WebsocketClient.Client)
	}
	client.groups[groupId][websocketId] = client.websocketClients[websocketId]
	client.websocketClientGroups[websocketId][groupId] = true
	return nil
}

func (client *Client) RemoveFromGroup(groupId string, websocketId string) error {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.groups[groupId] == nil {
		return Utilities.NewError("Group with id "+groupId+" does not exist", nil)
	}
	if client.groups[groupId][websocketId] == nil {
		return Utilities.NewError("WebsocketClient with id "+websocketId+" is not in group "+groupId, nil)
	}
	if client.websocketClients[websocketId] == nil {
		return Utilities.NewError("WebsocketClient with id "+websocketId+" does not exist", nil)
	}
	if client.websocketClientGroups[websocketId] == nil {
		return Utilities.NewError("WebsocketClient with id "+websocketId+" is not in any groups", nil)
	}
	delete(client.websocketClientGroups[websocketId], groupId)
	delete(client.groups[groupId], websocketId)
	if len(client.groups[groupId]) == 0 {
		delete(client.groups, groupId)
	}
	return nil
}

func (client *Client) GetGroupMembers(groupId string) []string {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.groups[groupId] == nil {
		return nil
	}
	groupMembers := make([]string, 0)
	for websocketId := range client.groups[groupId] {
		groupMembers = append(groupMembers, websocketId)
	}
	return groupMembers
}

func (client *Client) GetGroups(websocketId string) []string {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.websocketClients[websocketId] == nil {
		return nil
	}
	if client.websocketClientGroups[websocketId] == nil {
		return nil
	}
	groups := make([]string, 0)
	for groupId := range client.websocketClientGroups[websocketId] {
		groups = append(groups, groupId)
	}
	return groups
}

func (client *Client) IsInGroup(groupId string, websocketId string) bool {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.groups[groupId] == nil {
		return false
	}
	if client.groups[groupId][websocketId] == nil {
		return false
	}
	return true
}
