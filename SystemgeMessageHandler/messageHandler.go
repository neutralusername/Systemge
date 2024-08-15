package SystemgeMessageHandler

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

type SystemgeMessageHandler struct {
	asyncMessageHandlers map[string]func(*Message.Message)
	syncMessageHandlers  map[string]func(*Message.Message) (string, error)
	syncMutex            sync.Mutex
	asyncMutex           sync.Mutex
	sequentialMutex      sync.RWMutex

	// metrics
	asyncMessagesHandled atomic.Uint64
	syncRequestsHandled  atomic.Uint64
}

func NewMessageHandler(asyncMessageHandlers map[string]func(*Message.Message), syncMessageHandlers map[string]func(*Message.Message) (string, error)) *SystemgeMessageHandler {
	return &SystemgeMessageHandler{
		asyncMessageHandlers: asyncMessageHandlers,
		syncMessageHandlers:  syncMessageHandlers,
	}
}

func (messageHandler *SystemgeMessageHandler) HandleAsyncMessage(message *Message.Message) error {
	messageHandler.sequentialMutex.RLock()
	defer messageHandler.sequentialMutex.RUnlock()
	messageHandler.asyncMutex.Lock()
	handler, exists := messageHandler.asyncMessageHandlers[message.GetTopic()]
	messageHandler.asyncMutex.Unlock()
	if !exists {
		return Error.New("No handler for async message", nil)
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
		return "", Error.New("No handler for sync message", nil)
	}
	messageHandler.syncRequestsHandled.Add(1)
	return handler(message)
}
