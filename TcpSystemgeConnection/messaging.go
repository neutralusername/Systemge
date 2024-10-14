package TcpSystemgeConnection

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *TcpSystemgeConnection) write(messageBytes []byte, circumstance string) error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.WritingMessage,
		"sending message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: circumstance,
			Event.Bytes:        string(messageBytes),
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	bytesSent, err := Tcp.Write(connection.netConn, messageBytes, connection.config.TcpSendTimeoutMs)
	if err != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.WriteMessageFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: circumstance,
				Event.Bytes:        string(messageBytes),
			}),
		)
		if Tcp.IsConnectionClosed(err) {
			connection.Close()
		}
		return err
	}
	connection.bytesSent.Add(bytesSent)

	connection.onEvent(Event.NewInfoNoOption(
		Event.WroteMessage,
		"message sent",
		Event.Context{
			Event.Circumstance: circumstance,
			Event.Bytes:        string(messageBytes),
		},
	))

	return nil
}

func (connection *TcpSystemgeConnection) SyncResponse(message *Message.Message, success bool, payload string) error {

	if message == nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.UnexpectedNilValue,
			"received nil message in SyncResponse",
			Event.Context{
				Event.Circumstance: Event.SyncResponse,
				Event.Message:      string(message.Serialize()),
				Event.ResponseType: Event.Success,
				Event.Payload:      payload,
			},
		))
		return errors.New("received nil message in SyncResponse")
	}

	if message.GetSyncToken() == "" {
		connection.onEvent(Event.NewWarningNoOption(
			Event.NoSyncToken,
			"no sync token",
			Event.Context{
				Event.Circumstance: Event.SyncResponse,
				Event.Message:      string(message.Serialize()),
				Event.ResponseType: Event.Success,
				Event.Payload:      payload,
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

	if err := connection.write(response.Serialize(), Event.SyncResponse); err != nil {
		return err
	}
	connection.syncResponsesSent.Add(1)
	return nil
}

func (connection *TcpSystemgeConnection) AsyncMessage(topic, payload string) error {
	if err := connection.write(Message.NewAsync(topic, payload).Serialize(), Event.AsyncMessage); err != nil {
		return err
	}
	connection.asyncMessagesSent.Add(1)
	return nil
}

func (connection *TcpSystemgeConnection) SyncRequest(topic, payload string) (<-chan *Message.Message, error) {
	synctoken, syncRequestStruct := connection.initResponseChannel()

	if err := connection.write(Message.NewSync(topic, payload, synctoken).Serialize(), Event.AsyncMessage); err != nil {
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
	resChan, err := connection.SyncRequest(topic, payload)
	if err != nil {
		return nil, err
	}

	responseMessage, ok := <-resChan
	if !ok {
		return nil, errors.New("no response received")
	}
	return responseMessage, nil
}

func (connection *TcpSystemgeConnection) AbortSyncRequest(syncToken string) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.AbortingSyncRequest,
		"aborting sync request",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AbortSyncRequest,
			Event.SyncToken:    syncToken,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}
	syncRequestStruct, ok := connection.syncRequests[syncToken]
	if !ok {
		connection.onEvent(Event.NewWarningNoOption(
			Event.UnknownSyncToken,
			"no sync token",
			Event.Context{
				Event.Circumstance: Event.AbortSyncRequest,
				Event.SyncToken:    syncToken,
			},
		))
		return errors.New("no response channel found")
	}

	close(syncRequestStruct.abortChannel)
	delete(connection.syncRequests, syncToken)

	connection.onEvent(Event.NewInfoNoOption(
		Event.AbortedSyncRequest,
		"sync request aborted",
		Event.Context{
			Event.Circumstance: Event.AbortSyncRequest,
			Event.SyncToken:    syncToken,
		},
	))
	return nil
}

func (connection *TcpSystemgeConnection) initResponseChannel() (string, *syncRequestStruct) {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.InitializingSyncRequest,
		"initializing response channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SyncRequest,
		},
	)); !event.IsInfo() {
		return "", nil
	}

	syncToken := connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	for _, ok := connection.syncRequests[syncToken]; ok; {
		syncToken = connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	}
	connection.syncRequests[syncToken] = &syncRequestStruct{
		responseChannel: make(chan *Message.Message, 1),
		abortChannel:    make(chan bool),
	}

	if event := connection.onEvent(Event.NewInfoNoOption(
		Event.InitializedSyncRequest,
		"response channel initialized",
		Event.Context{
			Event.Circumstance: Event.SyncRequest,
			Event.SyncToken:    syncToken,
		},
	)); !event.IsInfo() {
		delete(connection.syncRequests, syncToken)
		return "", nil
	}

	return syncToken, connection.syncRequests[syncToken]
}

func (connection *TcpSystemgeConnection) addSyncResponse(message *Message.Message) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.AddingSyncResponse,
		"adding sync response",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AddSyncResponse,
			Event.Message:      string(message.Serialize()),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	syncRequestStruct, ok := connection.syncRequests[message.GetSyncToken()]
	if !ok {
		connection.onEvent(Event.NewWarningNoOption(
			Event.UnknownSyncToken,
			"unknown sync key",
			Event.Context{
				Event.Circumstance: Event.AddSyncResponse,
				Event.Message:      string(message.Serialize()),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		return errors.New("no response channel found")
	}

	syncRequestStruct.responseChannel <- message
	close(syncRequestStruct.responseChannel)
	delete(connection.syncRequests, message.GetSyncToken())
	if message.GetTopic() == Message.TOPIC_SUCCESS {
		connection.syncSuccessResponsesReceived.Add(1)
	} else if message.GetTopic() == Message.TOPIC_FAILURE {
		connection.syncFailureResponsesReceived.Add(1)
	}

	connection.onEvent(Event.NewInfoNoOption(
		Event.AddedSyncResponse,
		"sync response added",
		Event.Context{
			Event.Circumstance: Event.AddSyncResponse,
			Event.Address:      connection.GetAddress(),
			Event.Message:      string(message.Serialize()),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))
	return nil
}

func (connection *TcpSystemgeConnection) removeSyncRequest(syncToken string) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()

	connection.onEvent(Event.NewInfoNoOption(
		Event.RemovingSyncRequest,
		"removing sync request",
		Event.Context{
			Event.Circumstance: Event.RemoveSyncRequest,
			Event.SyncToken:    syncToken,
		},
	))

	_, ok := connection.syncRequests[syncToken]
	if !ok {
		connection.onEvent(Event.NewWarningNoOption(
			Event.UnknownSyncToken,
			"unknown sync key",
			Event.Context{
				Event.Circumstance: Event.RemoveSyncRequest,
				Event.SyncToken:    syncToken,
			},
		))
		return errors.New("no response channel found")
	}
	delete(connection.syncRequests, syncToken)

	connection.onEvent(Event.NewInfoNoOption(
		Event.RemovedSyncRequest,
		"sync request removed",
		Event.Context{
			Event.Circumstance: Event.RemoveSyncRequest,
			Event.SyncToken:    syncToken,
		},
	))

	return nil
}

type syncRequestStruct struct {
	responseChannel chan *Message.Message
	abortChannel    chan bool
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
