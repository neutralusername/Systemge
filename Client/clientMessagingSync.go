package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
)

// resolves the broker address for the provided topic and sends the sync message to the broker responsible for the topic and waits for a response.
func (client *Client) SyncMessage(topic, origin, payload string) (*Message.Message, error) {
	message := Message.NewSync(topic, origin, payload)
	if !client.isStarted {
		return nil, Utilities.NewError("Client not started", nil)
	}
	if message.GetSyncRequestToken() == "" {
		return nil, Utilities.NewError("SyncRequestToken not set", nil)
	}
	brokerConnection, err := client.getBrokerConnectionForTopic(message.GetTopic())
	if err != nil {
		return nil, Utilities.NewError("Error resolving broker address for topic \""+message.GetTopic()+"\"", err)
	}
	responseChannel, err := client.addMessageWaitingForResponse(message)
	if err != nil {
		return nil, Utilities.NewError("Error adding message to waiting for response map", err)
	}
	err = brokerConnection.send(message)
	if err != nil {
		client.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return nil, Utilities.NewError("Error sending sync request message", err)
	}
	return client.receiveSyncResponse(message, responseChannel)
}

func (client *Client) receiveSyncResponse(message *Message.Message, responseChannel chan *Message.Message) (*Message.Message, error) {
	select {
	case response := <-responseChannel:
		client.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		if response.GetTopic() == "error" {
			return nil, Utilities.NewError(response.GetPayload(), nil)
		}
		return response, nil
	case <-client.stopChannel:
		client.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return nil, Utilities.NewError("client stopped", nil)
	}
}
func (client *Client) addMessageWaitingForResponse(message *Message.Message) (chan *Message.Message, error) {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	if client.messagesWaitingForResponse[message.GetSyncRequestToken()] != nil {
		return nil, Utilities.NewError("SyncRequestToken already exists", nil)
	}
	responseChannel := make(chan *Message.Message)
	client.messagesWaitingForResponse[message.GetSyncRequestToken()] = responseChannel
	return responseChannel, nil
}
func (client *Client) removeMessageWaitingForResponse(syncRequestToken string, responseChannel chan *Message.Message) {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	close(responseChannel)
	delete(client.messagesWaitingForResponse, syncRequestToken)
}
