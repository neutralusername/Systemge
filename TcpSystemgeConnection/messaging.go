package TcpSystemgeConnection

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *TcpSystemgeConnection) SyncResponse(message *Message.Message, success bool, payload string) error {
	if event := connection.onEvent(Event.NewInfo(
		Event.SendingMessage,
		"sending sync response",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.SyncResponse,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.name,
			Event.ClientAddress: connection.GetAddress(),
			Event.Message:       string(message.Serialize()),
			Event.ResponseType:  Event.Success,
			Event.Payload:       payload,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if message == nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.UnexpectedNilValue,
			"received nil message in SyncResponse",
			Event.Context{
				Event.Circumstance:  Event.SyncResponse,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.name,
				Event.ClientAddress: connection.GetAddress(),
				Event.Message:       string(message.Serialize()),
				Event.ResponseType:  Event.Success,
				Event.Payload:       payload,
			},
		))
		return errors.New("received nil message in SyncResponse")
	}

	if message.GetSyncToken() == "" {
		connection.onEvent(Event.NewWarningNoOption(
			Event.NoSyncToken,
			"no sync token",
			Event.Context{
				Event.Circumstance:  Event.SyncResponse,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.name,
				Event.ClientAddress: connection.GetAddress(),
				Event.Message:       string(message.Serialize()),
				Event.ResponseType:  Event.Success,
				Event.Payload:       payload,
			},
		))
		return errors.New("no sync token")
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

	connection.onEvent(Event.NewInfoNoOption(
		Event.SentMessage,
		"sync response sent",
		Event.Context{
			Event.Circumstance:  Event.SyncResponse,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.name,
			Event.ClientAddress: connection.GetAddress(),
		},
	))

	return nil
}

func (connection *TcpSystemgeConnection) AsyncMessage(topic, payload string) error {

	if event := connection.onEvent(Event.NewInfo(
		Event.SendingMessage,
		"sending async message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.AsyncMessage,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.name,
			Event.ClientAddress: connection.GetAddress(),
			Event.Topic:         topic,
			Event.Payload:       payload,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	err := connection.send(Message.NewAsync(topic, payload).Serialize())
	if err != nil {
		return err
	}
	connection.asyncMessagesSent.Add(1)

	connection.onEvent(Event.NewInfoNoOption(
		Event.SentMessage,
		"async message sent",
		Event.Context{
			Event.Circumstance:  Event.AsyncMessage,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.name,
			Event.ClientAddress: connection.GetAddress(),
			Event.Topic:         topic,
			Event.Payload:       payload,
		},
	))
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
		return nil, Event.New("SyncRequest aborted", nil)

	case <-connection.closeChannel:
		connection.noSyncResponseReceived.Add(1)
		connection.removeSyncRequest(synctoken)
		return nil, Event.New("SystemgeClient stopped before receiving response", nil)

	case <-timeout:
		connection.noSyncResponseReceived.Add(1)
		connection.removeSyncRequest(synctoken)
		return nil, Event.New("Timeout before receiving response", nil)

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
	return Event.New("No response channel found", nil)
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

	syncRequestStruct, ok := connection.syncRequests[message.GetSyncToken()]
	if !ok {
		return errors.New("no response channel found")
	}

	syncRequestStruct.responseChannel <- message
	close(syncRequestStruct.responseChannel)
	delete(connection.syncRequests, message.GetSyncToken())

	return nil
}

func (connection *TcpSystemgeConnection) removeSyncRequest(syncToken string) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	if _, ok := connection.syncRequests[syncToken]; ok {
		delete(connection.syncRequests, syncToken)
		return nil
	}
	return Event.New("No response channel found", nil)
}

type syncRequestStruct struct {
	responseChannel chan *Message.Message
	abortChannel    chan bool
}

func (connection *TcpSystemgeConnection) send(bytes []byte) error {
	connection.sendMutex.Lock()
	defer connection.sendMutex.Unlock()
	bytesSent, err := Tcp.Send(connection.netConn, bytes, connection.config.TcpSendTimeoutMs)
	if err != nil {
		if Tcp.IsConnectionClosed(err) {
			connection.Close()
		}
		return err
	}
	connection.bytesSent.Add(bytesSent)
	return nil
}
