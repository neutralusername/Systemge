package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"time"
)

// resolves the broker address for the provided topic and sends the sync message to the broker responsible for the topic and waits for a response.
func (node *Node) SyncMessage(topic, origin, payload string) (*Message.Message, error) {
	message := Message.NewSync(topic, origin, payload, node.randomizer.GenerateRandomString(10, Utilities.ALPHA_NUMERIC))
	if !node.isStarted {
		return nil, Error.New("Node not started", nil)
	}
	if message.GetSyncRequestToken() == "" {
		return nil, Error.New("SyncRequestToken not set", nil)
	}
	brokerConnection, err := node.getBrokerConnectionForTopic(message.GetTopic())
	if err != nil {
		return nil, Error.New("Error resolving broker address for topic \""+message.GetTopic()+"\"", err)
	}
	responseChannel, err := node.addMessageWaitingForResponse(message)
	if err != nil {
		return nil, Error.New("Error adding message to waiting for response map", err)
	}
	err = brokerConnection.send(message)
	if err != nil {
		node.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return nil, Error.New("Error sending sync request message", err)
	}
	return node.receiveSyncResponse(message, responseChannel)
}

func (node *Node) receiveSyncResponse(message *Message.Message, responseChannel chan *Message.Message) (*Message.Message, error) {
	timeout := time.NewTimer(time.Duration(node.config.SyncMessageTimeoutMs) * time.Millisecond)
	defer func() {
		node.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		timeout.Stop()
	}()
	select {
	case response := <-responseChannel:
		if response.GetTopic() == "error" {
			return nil, Error.New(response.GetPayload(), nil)
		}
		return response, nil
	case <-node.stopChannel:
		return nil, Error.New("Node stopped", nil)
	case <-timeout.C:
		return nil, Error.New("Timeout waiting for response", nil)
	}
}
func (node *Node) addMessageWaitingForResponse(message *Message.Message) (chan *Message.Message, error) {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.messagesWaitingForResponse[message.GetSyncRequestToken()] != nil {
		return nil, Error.New("SyncRequestToken already exists", nil)
	}
	responseChannel := make(chan *Message.Message)
	node.messagesWaitingForResponse[message.GetSyncRequestToken()] = responseChannel
	return responseChannel, nil
}
func (node *Node) removeMessageWaitingForResponse(syncRequestToken string, responseChannel chan *Message.Message) {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	close(responseChannel)
	delete(node.messagesWaitingForResponse, syncRequestToken)
}
