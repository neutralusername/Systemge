package WebsocketServer

import (
	"Systemge/Message"
)

func (server *Server) Broadcast(message *Message.Message) {
	messageBytes := message.Serialize()
	server.acquireMutex()
	defer server.releaseMutex()
	for _, websocketClient := range server.clients {
		go websocketClient.Send(messageBytes)
	}
}

func (server *Server) Unicast(id string, message *Message.Message) {
	messageBytes := message.Serialize()
	server.acquireMutex()
	defer server.releaseMutex()
	if websocketClient, exists := server.clients[id]; exists {
		go websocketClient.Send(messageBytes)
	}
}

func (server *Server) Multicast(ids []string, message *Message.Message) {
	messageBytes := message.Serialize()
	server.acquireMutex()
	defer server.releaseMutex()
	for _, id := range ids {
		if websocketClient, exists := server.clients[id]; exists {
			go websocketClient.Send(messageBytes)
		}
	}
}

func (server *Server) Groupcast(groupId string, message *Message.Message) {
	messageBytes := message.Serialize()
	server.acquireMutex()
	defer server.releaseMutex()
	if server.groups[groupId] == nil {
		return
	}
	for _, websocketClient := range server.groups[groupId] {
		go websocketClient.Send(messageBytes)
	}
}
