package Client

import "Systemge/Error"

func (client *Client) AddToWebsocketGroup(groupId string, websocketId string) error {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	websocketClient := client.websocketClients[websocketId]
	if websocketClient == nil {
		return Error.New("WebsocketClient with id "+websocketId+" does not exist", nil)
	}
	if client.websocketClientGroups[websocketId][groupId] {
		return Error.New("WebsocketClient with id "+websocketId+" is already in group "+groupId, nil)
	}
	if client.WebsocketGroups[groupId] == nil {
		client.WebsocketGroups[groupId] = make(map[string]*WebsocketClient)
	}
	client.WebsocketGroups[groupId][websocketId] = client.websocketClients[websocketId]
	client.websocketClientGroups[websocketId][groupId] = true
	return nil
}

func (client *Client) RemoveFromWebsocketGroup(groupId string, websocketId string) error {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.WebsocketGroups[groupId] == nil {
		return Error.New("Group with id "+groupId+" does not exist", nil)
	}
	if client.WebsocketGroups[groupId][websocketId] == nil {
		return Error.New("WebsocketClient with id "+websocketId+" is not in group "+groupId, nil)
	}
	if client.websocketClients[websocketId] == nil {
		return Error.New("WebsocketClient with id "+websocketId+" does not exist", nil)
	}
	if client.websocketClientGroups[websocketId] == nil {
		return Error.New("WebsocketClient with id "+websocketId+" is not in any groups", nil)
	}
	delete(client.websocketClientGroups[websocketId], groupId)
	delete(client.WebsocketGroups[groupId], websocketId)
	if len(client.WebsocketGroups[groupId]) == 0 {
		delete(client.WebsocketGroups, groupId)
	}
	return nil
}

func (client *Client) GetWebsocketGroupClients(groupId string) []string {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.WebsocketGroups[groupId] == nil {
		return nil
	}
	groupMembers := make([]string, 0)
	for websocketId := range client.WebsocketGroups[groupId] {
		groupMembers = append(groupMembers, websocketId)
	}
	return groupMembers
}

func (client *Client) GetWebsocketGroups(websocketId string) []string {
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

func (client *Client) IsInWebsocketGroup(groupId string, websocketId string) bool {
	client.websocketMutex.Lock()
	defer client.websocketMutex.Unlock()
	if client.WebsocketGroups[groupId] == nil {
		return false
	}
	if client.WebsocketGroups[groupId][websocketId] == nil {
		return false
	}
	return true
}
