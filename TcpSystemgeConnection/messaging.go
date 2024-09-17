package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
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
	connection.syncResponsesSent.Add(1)
	return nil
}

func (connection *TcpSystemgeConnection) AsyncMessage(topic, payload string) error {
	err := connection.send(Message.New(topic, payload).Serialize())
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
