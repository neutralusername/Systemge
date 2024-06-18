package Broker

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"strings"
)

func (server *Server) handleBrokerClientMessages(clientConnection *clientConnection) {
	for server.isStarted {
		messageBytes, err := clientConnection.receive()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") && !strings.Contains(err.Error(), "EOF") {
				server.logger.Log(Utilities.NewError("Failed to receive message from client \""+clientConnection.name+"\"", err).Error())
			}
			clientConnection.disconnect()
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil || message.GetOrigin() == "" {
			server.logger.Log(Utilities.NewError("Invalid message from client \""+clientConnection.name+"\"", nil).Error())
			clientConnection.disconnect()
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
			errReponse := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("sync request failed", err).Error()))
			if errReponse != nil {
				return Utilities.NewError("Failed to handle sync request from client \""+clientConnection.name+"\" and failed to send error response", errReponse)
			}
			return Utilities.NewError("Failed to handle sync request from client \""+clientConnection.name+"\"", err)
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
			errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("", err).Error()))
			if errResponse != nil {
				return Utilities.NewError("Failed to unsubscribe client \""+clientConnection.name+"\" from topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
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
			errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("", err).Error()))
			if errResponse != nil {
				return Utilities.NewError("Failed to subscribe client \""+clientConnection.name+"\" to topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
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
			errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, "No message to consume"))
			if errResponse != nil {
				return Utilities.NewError("Failed to send error response to client \""+clientConnection.name+"\" and no message to consume", errResponse)
			}
			return Utilities.NewError("No message to consume for client \""+clientConnection.name+"\"", nil)
		}
		err := server.handleSyncResponse(message)
		if err != nil {
			return Utilities.NewError("Failed to send response to client \""+clientConnection.name+"\"", err)
		}
	default:
		server.operationMutex.Lock()
		clients := server.getSubscribers(message.GetTopic())
		server.operationMutex.Unlock()
		server.propagateMessage(clients, message)
	}
	return nil
}
