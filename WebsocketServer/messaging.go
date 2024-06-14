package WebsocketServer

import "Systemge/Message"

func (server *Server) Broadcast(message *Message.Message) {
	messageBytes := message.Serialize()
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, websocketClient := range server.clients {
		go websocketClient.Send(messageBytes)
	}
}

func (server *Server) Unicast(id string, message *Message.Message) {
	messageBytes := message.Serialize()
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if websocketClient, exists := server.clients[id]; exists {
		go websocketClient.Send(messageBytes)
	}
}

func (server *Server) Multicast(ids []string, message *Message.Message) {
	messageBytes := message.Serialize()
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, id := range ids {
		if websocketClient, exists := server.clients[id]; exists {
			go websocketClient.Send(messageBytes)
		}
	}
}

func (server *Server) Groupcast(groupId string, message *Message.Message) {
	messageBytes := message.Serialize()
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if server.groups[groupId] == nil {
		return
	}
	for _, websocketClient := range server.groups[groupId] {
		go websocketClient.Send(messageBytes)
	}
}
