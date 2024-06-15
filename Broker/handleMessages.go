package Broker

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"strings"
)

func (server *Server) handleClientMessages(clientConnection *clientConnection) {
	for server.IsStarted() {
		messageBytes, err := clientConnection.receive()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") && !strings.Contains(err.Error(), "EOF") {
				panic(err)
			} else {
				clientConnection.disconnect()
				return
			}
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			server.logger.Log(Utilities.NewError("Failed to deserialize message from client \""+clientConnection.name+"\"", nil).Error())
			return
		}
		if message.GetSyncResponseToken() != "" {
			err := server.handleSyncResponse(message)
			if err != nil {
				server.logger.Log(Utilities.NewError("Failed to handle sync response from client \""+clientConnection.name+"\"", err).Error())
			}
		} else {
			err = server.handleMessage(clientConnection, message)
			if err != nil {
				server.logger.Log(Utilities.NewError("Failed to handle message from client \""+clientConnection.name+"\"", err).Error())
			}
		}
	}
}

func (server *Server) handleMessage(clientConnection *clientConnection, message *Message.Message) error {
	err := server.validateMessageTopic(message)
	if err != nil {
		return Utilities.NewError("Failed to validate message topic", err)
	}
	if message.GetSyncRequestToken() != "" {
		if err := server.handleSyncRequest(clientConnection, message); err != nil {
			if err := clientConnection.send(message.NewResponse("error", server.name, Utilities.NewError("", err).Error())); err != nil {
				return Utilities.NewError("Failed to send error response to sync request from \""+clientConnection.name+"\"", err)
			}
		}
	}
	switch message.GetTopic() {
	case "heartbeat":
		err := clientConnection.resetWatchdog()
		if err != nil {
			return Utilities.NewError("Failed to reset watchdog for client \""+clientConnection.name+"\"", nil)
		}
		return nil
	case "unsubscribe":
		err := server.removeSubscription(clientConnection, message.GetPayload())
		if err != nil {
			err := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("", err).Error()))
			if err != nil {
				return Utilities.NewError("Failed to unsubscribe client \""+clientConnection.name+"\" from topic \""+message.GetPayload()+"\" and failed to send error response", err)
			}
			return Utilities.NewError("Failed to unsubscribe client \""+clientConnection.name+"\" from topic \""+message.GetPayload()+"\"", err)
		}
		err = server.handleSyncResponse(message.NewResponse("unsubscribed", server.name, ""))
		if err != nil {
			return Utilities.NewError("Failed to send unsubscribe response to client \""+clientConnection.name+"\"", err)
		}
		return nil
	case "subscribe":
		err := server.addSubscription(clientConnection, message.GetPayload())
		if err != nil {
			err := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("", err).Error()))
			if err != nil {
				return Utilities.NewError("Failed to subscribe client \""+clientConnection.name+"\" to topic \""+message.GetPayload()+"\" and failed to send error response", err)
			}
			return Utilities.NewError("Failed to subscribe client \""+clientConnection.name+"\" to topic \""+message.GetPayload()+"\"", err)
		}
		err = server.handleSyncResponse(message.NewResponse("subscribed", server.name, ""))
		if err != nil {
			return Utilities.NewError("Failed to send subscribe response to client \""+clientConnection.name+"\"", err)
		}
		return nil
	case "consume":
		message := clientConnection.dequeueMessage_Timeout(DEFAULT_TCP_TIMEOUT)
		if message == nil {
			err := server.handleSyncResponse(message.NewResponse("error", server.name, "No message to consume"))
			if err != nil {
				return Utilities.NewError("Failed to send error response to client \""+clientConnection.name+"\" and no message to consume", err)
			}
			return Utilities.NewError("No message to consume for client \""+clientConnection.name+"\"", nil)
		}
		err := server.handleSyncResponse(message)
		if err != nil {
			return Utilities.NewError("Failed to send response to client \""+clientConnection.name+"\"", err)
		}
	default:
		server.mutex.Lock()
		clients := server.getSubscribers(message.GetTopic())
		server.mutex.Unlock()
		server.propagateMessage(clients, message)
	}
	return nil
}

func (server *Server) handleSyncRequest(clientConnection *clientConnection, message *Message.Message) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if openSyncRequest := server.openSyncRequests[message.GetSyncRequestToken()]; openSyncRequest != nil {
		return Utilities.NewError("token already in use", nil)
	}
	if len(server.clientSubscriptions[message.GetTopic()]) == 0 && message.GetTopic() != "subscribe" && message.GetTopic() != "unsubscribe" {
		return Utilities.NewError("no subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	if len(server.clientSubscriptions[message.GetTopic()]) > 1 {
		return Utilities.NewError("multiple subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	server.openSyncRequests[message.GetSyncRequestToken()] = clientConnection
	clientConnection.openSyncRequests[message.GetSyncRequestToken()] = message
	return nil
}

func (server *Server) handleSyncResponse(message *Message.Message) error {
	server.mutex.Lock()
	waitingClientConnection := server.openSyncRequests[message.GetSyncResponseToken()]
	if waitingClientConnection == nil {
		server.mutex.Unlock()
		return Utilities.NewError("response to unknown sync request \""+message.GetSyncResponseToken()+"\"", nil)
	}
	delete(server.openSyncRequests, message.GetSyncResponseToken())
	delete(waitingClientConnection.openSyncRequests, message.GetSyncResponseToken())
	server.mutex.Unlock()
	err := waitingClientConnection.send(message)
	if err != nil {
		return Utilities.NewError("Failed to send response to client \""+waitingClientConnection.name+"\"", err)
	}
	return nil
}
