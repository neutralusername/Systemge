package SystemgeMessageHandler

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type MessageHandler interface {
	HandleAsyncMessage(connection SystemgeConnection.SystemgeConnection, message *Message.Message) error
	HandleSyncRequest(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error)
	AddAsyncMessageHandler(topic string, handler AsyncMessageHandler)
	AddSyncMessageHandler(topic string, handler SyncMessageHandler)
	RemoveAsyncMessageHandler(topic string)
	RemoveSyncMessageHandler(topic string)
	SetUnknownAsyncHandler(handler AsyncMessageHandler)
	SetUnknownSyncHandler(handler SyncMessageHandler)
	GetAsyncMessageHandler(topic string) AsyncMessageHandler
	GetSyncMessageHandler(topic string) SyncMessageHandler
	GetAsyncTopics() []string
	GetSyncTopics() []string

	CheckMetrics() map[string]*Metrics.Metrics
	GetMetrics() map[string]*Metrics.Metrics

	CheckAsyncMessagesHandled() uint64
	GetAsyncMessagesHandled() uint64

	CheckSyncRequestsHandled() uint64
	GetSyncRequestsHandled() uint64

	CheckUnknownTopicsReceived() uint64
	GetUnknownTopicsReceived() uint64
}

type AsyncMessageHandler func(SystemgeConnection.SystemgeConnection, *Message.Message)
type AsyncMessageHandlers map[string]AsyncMessageHandler

type SyncMessageHandler func(SystemgeConnection.SystemgeConnection, *Message.Message) (string, error)
type SyncMessageHandlers map[string]SyncMessageHandler

func NewAsyncMessageHandlers() AsyncMessageHandlers {
	return make(AsyncMessageHandlers)
}

func (map1 AsyncMessageHandlers) Merge(map2 AsyncMessageHandlers) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 AsyncMessageHandlers) Add(key string, value AsyncMessageHandler) {
	map1[key] = value
}

func (map1 AsyncMessageHandlers) Remove(key string) {
	delete(map1, key)
}

func (map1 AsyncMessageHandlers) Get(key string) (AsyncMessageHandler, bool) {
	value, ok := map1[key]
	return value, ok
}

func NewSyncMessageHandlers() SyncMessageHandlers {
	return make(SyncMessageHandlers)
}

func (map1 SyncMessageHandlers) Merge(map2 SyncMessageHandlers) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 SyncMessageHandlers) Add(key string, value SyncMessageHandler) {
	map1[key] = value
}

func (map1 SyncMessageHandlers) Remove(key string) {
	delete(map1, key)
}

func (map1 SyncMessageHandlers) Get(key string) (SyncMessageHandler, bool) {
	value, ok := map1[key]
	return value, ok
}

// A started loop will run indefinitely until StopProcessingLoop is called.
func StartMessageHandlingLoop_Sequentially(connection SystemgeConnection.SystemgeConnection, messageHandler MessageHandler) (chan<- bool, <-chan error) {
	stopChannel := make(chan bool)
	errChannel := make(chan error)
	go func() {
		for {
			select {
			case <-stopChannel:
				errChannel <- Error.New("processing loop stopped", nil)
				close(errChannel)
				return
			default:
				message, err := connection.GetNextMessage()
				if err != nil {
					errChannel <- Error.New("failed to get next message", err)
					close(errChannel)
					return
				}
				if err := HandleMessage(connection, message, messageHandler); err != nil {
					errChannel <- Error.New("failed to process message", err)
				}
			}
		}
	}()
	return stopChannel, errChannel
}

// A started loop will run indefinitely until StopProcessingLoop is called.
func StartMessageHandlingLoop_Concurrently(connection SystemgeConnection.SystemgeConnection, messageHandler MessageHandler) (chan<- bool, <-chan error) {
	stopChannel := make(chan bool)
	errChannel := make(chan error)
	go func() {
		for {
			select {
			case <-stopChannel:
				errChannel <- Error.New("processing loop stopped", nil)
				close(errChannel)
				return
			default:
				message, err := connection.GetNextMessage()
				if err != nil {
					errChannel <- Error.New("failed to get next message", err)
					close(errChannel)
					return
				}
				go func() {
					if err := HandleMessage(connection, message, messageHandler); err != nil {
						errChannel <- Error.New("failed to process message", err)
					}
				}()
			}
		}
	}()
	return stopChannel, errChannel
}

func HandleMessage(connection SystemgeConnection.SystemgeConnection, message *Message.Message, messageHandler MessageHandler) error {
	if messageHandler == nil {
		return Error.New("no message handler set", nil)
	}
	if message.GetSyncToken() == "" {
		err := messageHandler.HandleAsyncMessage(connection, message)
		if err != nil {
			return Error.New("failed to handle async message", err)
		}
	} else {
		if responsePayload, err := messageHandler.HandleSyncRequest(connection, message); err != nil {
			if err := connection.SyncResponse(message, false, err.Error()); err != nil {
				return Error.New("failed to send failure response", err)
			}
		} else {
			if err := connection.SyncResponse(message, true, responsePayload); err != nil {
				return Error.New("failed to send success response", err)
			}
		}
	}
	return nil
}
