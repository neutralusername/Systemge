package SystemgeMessageHandler

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

type AsyncMessageHandlers map[string]func(*Message.Message)
type SyncMessageHandlers map[string]func(*Message.Message) (string, error)

type SystemgeMessageHandler struct {
	asyncMessageHandlers AsyncMessageHandlers
	syncMessageHandlers  SyncMessageHandlers
	syncMutex            sync.Mutex
	asyncMutex           sync.Mutex
	sequentialMutex      sync.RWMutex

	unknwonAsyncTopicHandler func(*Message.Message)
	unknwonSyncTopicHandler  func(*Message.Message) (string, error)

	// metrics
	asyncMessagesHandled atomic.Uint64
	syncRequestsHandled  atomic.Uint64
}

// pass a handler with an empty string as the key to handle messages with unknown topics
func New(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers) *SystemgeMessageHandler {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandlers := &SystemgeMessageHandler{
		asyncMessageHandlers: asyncMessageHandlers,
		syncMessageHandlers:  syncMessageHandlers,
	}
	if handler, exists := asyncMessageHandlers[""]; exists {
		systemgeMessageHandlers.unknwonAsyncTopicHandler = handler
	}
	if handler, exists := syncMessageHandlers[""]; exists {
		systemgeMessageHandlers.unknwonSyncTopicHandler = handler
	}
	return systemgeMessageHandlers
}

func (messageHandler *SystemgeMessageHandler) HandleAsyncMessage(message *Message.Message) error {
	messageHandler.sequentialMutex.RLock()
	defer messageHandler.sequentialMutex.RUnlock()
	messageHandler.asyncMutex.Lock()
	handler, exists := messageHandler.asyncMessageHandlers[message.GetTopic()]
	messageHandler.asyncMutex.Unlock()
	if !exists {
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
	messageHandler.sequentialMutex.RLock()
	defer messageHandler.sequentialMutex.RUnlock()
	messageHandler.syncMutex.Lock()
	handler, exists := messageHandler.syncMessageHandlers[message.GetTopic()]
	messageHandler.syncMutex.Unlock()
	if !exists {
		if messageHandler.unknwonSyncTopicHandler != nil {
			messageHandler.syncRequestsHandled.Add(1)
			return messageHandler.unknwonSyncTopicHandler(message)
		}
		return "", Error.New("No handler for sync message", nil)
	}
	messageHandler.syncRequestsHandled.Add(1)
	return handler(message)
}

func (messageHandler *SystemgeMessageHandler) HandleAsyncMessageSequentially(message *Message.Message) error {
	messageHandler.sequentialMutex.Lock()
	defer messageHandler.sequentialMutex.Unlock()
	handler, exists := messageHandler.asyncMessageHandlers[message.GetTopic()]
	if !exists {
		if messageHandler.unknwonAsyncTopicHandler != nil {
			messageHandler.asyncMessagesHandled.Add(1)
			messageHandler.unknwonAsyncTopicHandler(message)
			return nil
		}
		return Error.New("No handler for async message", nil)
	}
	messageHandler.asyncMessagesHandled.Add(1)
	handler(message)
	return nil
}

func (messageHandler *SystemgeMessageHandler) HandleSyncRequestSequentially(message *Message.Message) (string, error) {
	messageHandler.sequentialMutex.Lock()
	defer messageHandler.sequentialMutex.Unlock()
	handler, exists := messageHandler.syncMessageHandlers[message.GetTopic()]
	if !exists {
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
