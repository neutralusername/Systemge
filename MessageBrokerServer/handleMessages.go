package MessageBrokerServer

import (
	"Systemge/Error"
	"Systemge/Message"
	"strings"
)

func (server *Server) handleClientMessages(client *Client) {
	for server.IsStarted() {
		messageBytes, err := client.receive()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") && !strings.Contains(err.Error(), "EOF") {
				panic(err)
			} else {
				client.disconnect()
				return
			}
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			server.logger.Log(Error.New("Failed to deserialize message from client \""+client.name+"\"", nil).Error())
			return
		}
		if message.GetSyncResponseToken() != "" {
			err := server.handleSyncResponse(message)
			if err != nil {
				server.logger.Log(Error.New("Failed to handle sync response from client \""+client.name+"\"", err).Error())
			}
		} else {
			err = server.handleMessage(client, message)
			if err != nil {
				server.logger.Log(Error.New("Failed to handle message from client \""+client.name+"\"", err).Error())
			}
		}
	}
}

func (server *Server) handleMessage(client *Client, message *Message.Message) error {
	err := server.validateMessageTopic(message)
	if err != nil {
		return Error.New("Failed to validate message topic", err)
	}
	if message.GetSyncRequestToken() != "" {
		if err := server.handleSyncRequest(client, message); err != nil {
			if err := client.send(message.NewResponse("error", server.name, Error.New("", err).Error())); err != nil {
				return Error.New("Failed to send error response to sync request from \""+client.name+"\"", err)
			}
		}
	}
	switch message.GetTopic() {
	case "heartbeat":
		err := client.resetWatchdog()
		if err != nil {
			return Error.New("Failed to reset watchdog for client \""+client.name+"\"", nil)
		}
		return nil
	case "unsubscribe":
		err := server.removeSubscription(client, message.GetPayload())
		if err != nil {
			err := server.handleSyncResponse(message.NewResponse("error", server.name, Error.New("", err).Error()))
			if err != nil {
				return Error.New("Failed to unsubscribe client \""+client.name+"\" from topic \""+message.GetPayload()+"\" and failed to send error response", err)
			}
			return Error.New("Failed to unsubscribe client \""+client.name+"\" from topic \""+message.GetPayload()+"\"", err)
		}
		err = server.handleSyncResponse(message.NewResponse("unsubscribed", server.name, ""))
		if err != nil {
			return Error.New("Failed to send unsubscribe response to client \""+client.name+"\"", err)
		}
		return nil
	case "subscribe":
		err := server.addSubscription(client, message.GetPayload())
		if err != nil {
			err := server.handleSyncResponse(message.NewResponse("error", server.name, Error.New("", err).Error()))
			if err != nil {
				return Error.New("Failed to subscribe client \""+client.name+"\" to topic \""+message.GetPayload()+"\" and failed to send error response", err)
			}
			return Error.New("Failed to subscribe client \""+client.name+"\" to topic \""+message.GetPayload()+"\"", err)
		}
		err = server.handleSyncResponse(message.NewResponse("subscribed", server.name, ""))
		if err != nil {
			return Error.New("Failed to send subscribe response to client \""+client.name+"\"", err)
		}
		return nil
	case "consume":
		message := client.dequeueMessage_Timeout(DEFAULT_TCP_TIMEOUT)
		if message == nil {
			err := server.handleSyncResponse(message.NewResponse("error", server.name, "No message to consume"))
			if err != nil {
				return Error.New("Failed to send error response to client \""+client.name+"\" and no message to consume", err)
			}
			return Error.New("No message to consume for client \""+client.name+"\"", nil)
		}
		err := server.handleSyncResponse(message)
		if err != nil {
			return Error.New("Failed to send response to client \""+client.name+"\"", err)
		}
	default:
		server.mutex.Lock()
		clients := server.getSubscribers(message.GetTopic())
		server.mutex.Unlock()
		server.propagateMessage(clients, message)
	}
	return nil
}

func (server *Server) handleSyncRequest(client *Client, message *Message.Message) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if openSyncRequest := server.openSyncRequests[message.GetSyncRequestToken()]; openSyncRequest != nil {
		return Error.New("token already in use", nil)
	}
	if len(server.subscriptions[message.GetTopic()]) == 0 && message.GetTopic() != "subscribe" && message.GetTopic() != "unsubscribe" {
		return Error.New("no subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	if len(server.subscriptions[message.GetTopic()]) > 1 {
		return Error.New("multiple subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	server.openSyncRequests[message.GetSyncRequestToken()] = client
	client.openSyncRequests[message.GetSyncRequestToken()] = message
	return nil
}

func (server *Server) handleSyncResponse(message *Message.Message) error {
	server.mutex.Lock()
	waitingClient := server.openSyncRequests[message.GetSyncResponseToken()]
	if waitingClient == nil {
		server.mutex.Unlock()
		return Error.New("response to unknown sync request \""+message.GetSyncResponseToken()+"\"", nil)
	}
	delete(server.openSyncRequests, message.GetSyncResponseToken())
	delete(waitingClient.openSyncRequests, message.GetSyncResponseToken())
	server.mutex.Unlock()
	err := waitingClient.send(message)
	if err != nil {
		return Error.New("Failed to send response to client \""+waitingClient.name+"\"", err)
	}
	return nil
}
