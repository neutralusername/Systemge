package Node

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

// resolves the broker address for the provided topic and sends the sync message to the broker responsible for the topic and waits for a response.
func (node *Node) SyncMessage(topic, origin, payload string) (*Message.Message, error) {
	if systemge := node.systemge; systemge != nil {
		message := Message.NewSync(topic, origin, payload, node.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC))
		if !node.IsStarted() {
			return nil, Error.New("node not started", nil)
		}
		if message.GetSyncRequestToken() == "" {
			return nil, Error.New("syncRequestToken not set", nil)
		}
		brokerConnection, err := node.getBrokerConnectionForTopic(message.GetTopic(), true)
		if err != nil {
			return nil, Error.New("failed getting broker connection", err)
		}
		responseChannel, err := systemge.addMessageWaitingForResponse(message)
		if err != nil {
			return nil, Error.New("failed to add message waiting for response", err)
		}
		bytesSent, err := brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message.Serialize())
		if err != nil {
			systemge.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
			return nil, Error.New("failed sending sync request message", err)
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Sent sync request message with topic \""+message.GetTopic()+"\" to broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
		}
		systemge.bytesSentCounter.Add(bytesSent)
		systemge.outgoingSyncRequestCounter.Add(1)
		response, err := systemge.receiveSyncResponse(message, responseChannel)
		if err != nil {
			return nil, Error.New("failed receiving sync response", err)
		}
		systemge.incomingSyncResponseCounter.Add(1)
		return response, nil
	}
	return nil, Error.New("systemge not initialized", nil)
}

func (systemge *systemgeComponent) receiveSyncResponse(message *Message.Message, responseChannel chan *Message.Message) (*Message.Message, error) {
	timeout := time.NewTimer(time.Duration(systemge.application.GetSystemgeComponentConfig().SyncResponseTimeoutMs) * time.Millisecond)
	defer func() {
		systemge.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		timeout.Stop()
	}()
	select {
	case response := <-responseChannel:
		if response == nil {
			return nil, Error.New("response channel closed", nil)
		}
		if response.GetTopic() == "error" {
			return nil, Error.New(response.GetPayload(), nil)
		}
		return response, nil
	case <-timeout.C:
		return nil, Error.New("timeout waiting for response", nil)
	}
}
func (systemge *systemgeComponent) addMessageWaitingForResponse(message *Message.Message) (chan *Message.Message, error) {
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	if systemge.messagesWaitingForResponse[message.GetSyncRequestToken()] != nil {
		return nil, Error.New("syncRequestToken already exists", nil)
	}
	responseChannel := make(chan *Message.Message)
	systemge.messagesWaitingForResponse[message.GetSyncRequestToken()] = responseChannel
	return responseChannel, nil
}
func (systemge *systemgeComponent) removeMessageWaitingForResponse(syncRequestToken string, responseChannel chan *Message.Message) {
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	close(responseChannel)
	delete(systemge.messagesWaitingForResponse, syncRequestToken)
}
