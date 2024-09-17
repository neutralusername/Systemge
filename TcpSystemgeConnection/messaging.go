package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *TcpSystemgeConnection) SyncResponse(message *Message.Message, success bool, payload string) error {
	if message == nil {
		return Error.New("Message is nil", nil)
	}
	if message.GetSyncToken() == "" {
		return Error.New("SyncToken is empty", nil)
	}
	var response *Message.Message
	if success {
		response = message.NewSuccessResponse(payload)
	} else {
		response = message.NewFailureResponse(payload)
	}
	err := connection.send(response.Serialize())
	if err != nil {
		return err
	}
	connection.asyncMessagesSent.Add(1)
	return nil
}

func (connection *TcpSystemgeConnection) AsyncMessage(topic, payload string) error {
	err := connection.send(Message.NewAsync(topic, payload).Serialize())
	if err != nil {
		return err
	}
	connection.asyncMessagesSent.Add(1)
	return nil
}

// blocks until the sending attempt is completed. returns error if sending request fails.
// nil result from response channel indicates either connection closed before receiving response, timeout or manual abortion.
func (connection *TcpSystemgeConnection) SyncRequest(topic, payload string) (<-chan *Message.Message, error) {
	synctoken, syncRequestStruct := connection.initResponseChannel()
	err := connection.send(Message.NewSync(topic, payload, synctoken).Serialize())
	if err != nil {
		connection.removeSyncRequest(synctoken)
		return nil, err
	}
	connection.syncRequestsSent.Add(1)

	var timeout <-chan time.Time
	if connection.config.SyncRequestTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.SyncRequestTimeoutMs) * time.Millisecond)
	}

	resChan := make(chan *Message.Message, 1)
	go func() {
		select {
		case responseMessage := <-syncRequestStruct.responseChannel:
			if responseMessage.GetTopic() == Message.TOPIC_SUCCESS {
				connection.syncSuccessResponsesReceived.Add(1)
			} else if responseMessage.GetTopic() == Message.TOPIC_FAILURE {
				connection.syncFailureResponsesReceived.Add(1)
			}
			resChan <- responseMessage
			close(resChan)

		case <-syncRequestStruct.abortChannel:
			connection.noSyncResponseReceived.Add(1)
			close(resChan)

		case <-connection.closeChannel:
			connection.noSyncResponseReceived.Add(1)
			connection.removeSyncRequest(synctoken)
			close(resChan)

		case <-timeout:
			connection.noSyncResponseReceived.Add(1)
			connection.removeSyncRequest(synctoken)
			close(resChan)
		}
	}()
	return resChan, nil
}

// blocks until response is received, connection is closed, timeout or manual abortion.
func (connection *TcpSystemgeConnection) SyncRequestBlocking(topic, payload string) (*Message.Message, error) {
	synctoken, syncRequestStruct := connection.initResponseChannel()
	err := connection.send(Message.NewSync(topic, payload, synctoken).Serialize())
	if err != nil {
		connection.removeSyncRequest(synctoken)
		return nil, err
	}
	connection.syncRequestsSent.Add(1)

	var timeout <-chan time.Time
	if connection.config.SyncRequestTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.SyncRequestTimeoutMs) * time.Millisecond)
	}
	select {
	case responseMessage := <-syncRequestStruct.responseChannel:
		if responseMessage.GetTopic() == Message.TOPIC_SUCCESS {
			connection.syncSuccessResponsesReceived.Add(1)
		} else if responseMessage.GetTopic() == Message.TOPIC_FAILURE {
			connection.syncFailureResponsesReceived.Add(1)
		}
		return responseMessage, nil

	case <-syncRequestStruct.abortChannel:
		connection.noSyncResponseReceived.Add(1)
		return nil, Error.New("SyncRequest aborted", nil)

	case <-connection.closeChannel:
		connection.noSyncResponseReceived.Add(1)
		connection.removeSyncRequest(synctoken)
		return nil, Error.New("SystemgeClient stopped before receiving response", nil)

	case <-timeout:
		connection.noSyncResponseReceived.Add(1)
		connection.removeSyncRequest(synctoken)
		return nil, Error.New("Timeout before receiving response", nil)

	}
}

func (connection *TcpSystemgeConnection) AbortSyncRequest(syncToken string) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	if syncRequestStruct, ok := connection.syncRequests[syncToken]; ok {
		close(syncRequestStruct.abortChannel)
		delete(connection.syncRequests, syncToken)
		return nil
	}
	return Error.New("No response channel found", nil)
}

// returns a slice of syncTokens of open sync requests
func (connection *TcpSystemgeConnection) GetOpenSyncRequests() []string {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	syncTokens := make([]string, 0, len(connection.syncRequests))
	for k := range connection.syncRequests {
		syncTokens = append(syncTokens, k)
	}
	return syncTokens
}

func (connection *TcpSystemgeConnection) initResponseChannel() (string, *syncRequestStruct) {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	syncToken := connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	for _, ok := connection.syncRequests[syncToken]; ok; {
		syncToken = connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	}
	connection.syncRequests[syncToken] = &syncRequestStruct{
		responseChannel: make(chan *Message.Message, 1),
		abortChannel:    make(chan bool),
	}
	return syncToken, connection.syncRequests[syncToken]
}

func (connection *TcpSystemgeConnection) addSyncResponse(message *Message.Message) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	if syncRequestStruct, ok := connection.syncRequests[message.GetSyncToken()]; ok {
		syncRequestStruct.responseChannel <- message
		close(syncRequestStruct.responseChannel)
		delete(connection.syncRequests, message.GetSyncToken())
		return nil
	}
	return Error.New("No response channel found", nil)
}

func (connection *TcpSystemgeConnection) removeSyncRequest(syncToken string) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	if _, ok := connection.syncRequests[syncToken]; ok {
		delete(connection.syncRequests, syncToken)
		return nil
	}
	return Error.New("No response channel found", nil)
}

type syncRequestStruct struct {
	responseChannel chan *Message.Message
	abortChannel    chan bool
}
