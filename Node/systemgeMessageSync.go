package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tools"
	"time"
)

// resolves the broker address for the provided topic and sends the sync message to the broker responsible for the topic and waits for a response.
func (node *Node) SyncMessage(topic, origin, payload string) (*Message.Message, error) {
	if !node.systemgeStarted {
		return nil, Error.New("systemge component not started", nil)
	}
	message := Message.NewSync(topic, origin, payload, node.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC))
	if !node.IsStarted() {
		return nil, Error.New("node not started", nil)
	}
	if message.GetSyncRequestToken() == "" {
		return nil, Error.New("syncRequestToken not set", nil)
	}
	brokerConnection := node.getBrokerConnectionForTopic(message.GetTopic())
	if brokerConnection == nil {
		return nil, Error.New("failed resolving broker address for topic \""+message.GetTopic()+"\"", nil)
	}
	responseChannel, err := node.addMessageWaitingForResponse(message)
	if err != nil {
		return nil, Error.New("failed to add message waiting for response", err)
	}
	err = node.send(brokerConnection, message)
	if err != nil {
		node.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return nil, Error.New("failed sending sync request message", err)
	}
	response, err := node.receiveSyncResponse(message, responseChannel)
	if err != nil {
		return nil, Error.New("failed receiving sync response", err)
	}
	return response, nil
}

func (node *Node) receiveSyncResponse(message *Message.Message, responseChannel chan *Message.Message) (*Message.Message, error) {
	timeout := time.NewTimer(time.Duration(node.GetSystemgeComponent().GetSystemgeComponentConfig().SyncResponseTimeoutMs) * time.Millisecond)
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
		return nil, Error.New("node stopped", nil)
	case <-timeout.C:
		return nil, Error.New("timeout waiting for response", nil)
	}
}
func (node *Node) addMessageWaitingForResponse(message *Message.Message) (chan *Message.Message, error) {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	if node.systemgeMessagesWaitingForResponse[message.GetSyncRequestToken()] != nil {
		return nil, Error.New("syncRequestToken already exists", nil)
	}
	responseChannel := make(chan *Message.Message)
	node.systemgeMessagesWaitingForResponse[message.GetSyncRequestToken()] = responseChannel
	return responseChannel, nil
}
func (node *Node) removeMessageWaitingForResponse(syncRequestToken string, responseChannel chan *Message.Message) {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	close(responseChannel)
	delete(node.systemgeMessagesWaitingForResponse, syncRequestToken)
}
