package SystemgeMessageHandler

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

type AsyncMessageHandler func(*Message.Message)
type AsyncMessageHandlers map[string]AsyncMessageHandler

type SyncMessageHandler func(*Message.Message) (string, error)
type SyncMessageHandlers map[string]SyncMessageHandler

type SystemgeMessageHandler struct {
	asyncMessageHandlers AsyncMessageHandlers
	syncMessageHandlers  SyncMessageHandlers
	syncMutex            sync.Mutex
	asyncMutex           sync.Mutex
	sequentialMutex      sync.Mutex
	handleSequentially   bool

	unknwonAsyncTopicHandler func(*Message.Message)
	unknwonSyncTopicHandler  func(*Message.Message) (string, error)

	// metrics
	asyncMessagesHandled  atomic.Uint64
	syncRequestsHandled   atomic.Uint64
	unknownTopicsReceived atomic.Uint64
}

// handle sequentially has the effect that only one message is handled at a time
func New(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers, unknownTopicAsyncHandler *AsyncMessageHandler, unknownTopicSyncHandler *SyncMessageHandler, handleSequentially bool) *SystemgeMessageHandler {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandler := &SystemgeMessageHandler{
		asyncMessageHandlers:     asyncMessageHandlers,
		syncMessageHandlers:      syncMessageHandlers,
		handleSequentially:       handleSequentially,
		unknwonAsyncTopicHandler: *unknownTopicAsyncHandler,
		unknwonSyncTopicHandler:  *unknownTopicSyncHandler,
	}
	return systemgeMessageHandler
}

func (messageHandler *SystemgeMessageHandler) HandleAsyncMessage(message *Message.Message) error {
	if messageHandler.handleSequentially {
		messageHandler.sequentialMutex.Lock()
		defer messageHandler.sequentialMutex.Unlock()
	}
	messageHandler.asyncMutex.Lock()
	handler, exists := messageHandler.asyncMessageHandlers[message.GetTopic()]
	messageHandler.asyncMutex.Unlock()
	if !exists {
		messageHandler.unknownTopicsReceived.Add(1)
		if messageHandler.unknwonAsyncTopicHandler != nil {
			messageHandler.asyncMessagesHandled.Add(1)
			messageHandler.unknwonAsyncTopicHandler(message)
			return nil
		} else {
			return Error.New("No handler for async message", nil)
		}
	}
	messageHandler.asyncMessagesHandled.Add(1)
	handler(message)
	return nil
}

func (messageHandler *SystemgeMessageHandler) HandleSyncRequest(message *Message.Message) (string, error) {
	if messageHandler.handleSequentially {
		messageHandler.sequentialMutex.Lock()
		defer messageHandler.sequentialMutex.Unlock()
	}
	messageHandler.syncMutex.Lock()
	handler, exists := messageHandler.syncMessageHandlers[message.GetTopic()]
	messageHandler.syncMutex.Unlock()
	if !exists {
		messageHandler.unknownTopicsReceived.Add(1)
		if messageHandler.unknwonSyncTopicHandler != nil {
			messageHandler.syncRequestsHandled.Add(1)
			return messageHandler.unknwonSyncTopicHandler(message)
		}
		return "", Error.New("No handler for sync message", nil)
	}
	messageHandler.syncRequestsHandled.Add(1)
	return handler(message)
}

func (messageHandler *SystemgeMessageHandler) AddAsyncMessageHandler(topic string, handler func(*Message.Message)) {
	messageHandler.asyncMutex.Lock()
	messageHandler.asyncMessageHandlers[topic] = handler
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *SystemgeMessageHandler) AddSyncMessageHandler(topic string, handler func(*Message.Message) (string, error)) {
	messageHandler.syncMutex.Lock()
	messageHandler.syncMessageHandlers[topic] = handler
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *SystemgeMessageHandler) RemoveAsyncMessageHandler(topic string) {
	messageHandler.asyncMutex.Lock()
	delete(messageHandler.asyncMessageHandlers, topic)
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *SystemgeMessageHandler) RemoveSyncMessageHandler(topic string) {
	messageHandler.syncMutex.Lock()
	delete(messageHandler.syncMessageHandlers, topic)
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *SystemgeMessageHandler) SetUnknownAsyncHandler(handler func(*Message.Message)) {
	messageHandler.unknwonAsyncTopicHandler = handler
}

func (messageHandler *SystemgeMessageHandler) SetUnknownSyncHandler(handler func(*Message.Message) (string, error)) {
	messageHandler.unknwonSyncTopicHandler = handler
}

func (messageHandler *SystemgeMessageHandler) SetHandleSequentially(handleSequentially bool) {
	messageHandler.handleSequentially = handleSequentially
}

func (messageHandler *SystemgeMessageHandler) GetAsyncMessageHandler(topic string) func(*Message.Message) {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	return messageHandler.asyncMessageHandlers[topic]
}

func (messageHandler *SystemgeMessageHandler) GetSyncMessageHandler(topic string) func(*Message.Message) (string, error) {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	return messageHandler.syncMessageHandlers[topic]
}
