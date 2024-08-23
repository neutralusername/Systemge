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

type MessageHandler interface {
	HandleAsyncMessage(message *Message.Message) error
	HandleSyncRequest(message *Message.Message) (string, error)
	AddAsyncMessageHandler(topic string, handler func(*Message.Message))
	AddSyncMessageHandler(topic string, handler func(*Message.Message) (string, error))
	RemoveAsyncMessageHandler(topic string)
	RemoveSyncMessageHandler(topic string)
	SetUnknownAsyncHandler(handler func(*Message.Message))
	SetUnknownSyncHandler(handler func(*Message.Message) (string, error))
	GetAsyncMessageHandler(topic string) func(*Message.Message)
	GetSyncMessageHandler(topic string) func(*Message.Message) (string, error)

	GetAsyncMessagesHandled() uint64
	RetrieveAsyncMessagesHandled() uint64

	GetSyncRequestsHandled() uint64
	RetrieveSyncRequestsHandled() uint64

	GetUnknownTopicsReceived() uint64
	RetrieveUnknownTopicsReceived() uint64
}

type ConcurrentMessageHandler struct {
	asyncMessageHandlers AsyncMessageHandlers
	syncMessageHandlers  SyncMessageHandlers
	syncMutex            sync.Mutex
	asyncMutex           sync.Mutex

	unknwonAsyncTopicHandler func(*Message.Message)
	unknwonSyncTopicHandler  func(*Message.Message) (string, error)

	// metrics
	asyncMessagesHandled  atomic.Uint64
	syncRequestsHandled   atomic.Uint64
	unknownTopicsReceived atomic.Uint64
}

func NewConcurrentMessageHandler(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers, unknownTopicAsyncHandler *AsyncMessageHandler, unknownTopicSyncHandler *SyncMessageHandler) *ConcurrentMessageHandler {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandler := &ConcurrentMessageHandler{
		asyncMessageHandlers:     asyncMessageHandlers,
		syncMessageHandlers:      syncMessageHandlers,
		unknwonAsyncTopicHandler: *unknownTopicAsyncHandler,
		unknwonSyncTopicHandler:  *unknownTopicSyncHandler,
	}
	return systemgeMessageHandler
}

func (messageHandler *ConcurrentMessageHandler) HandleAsyncMessage(message *Message.Message) error {
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

func (messageHandler *ConcurrentMessageHandler) HandleSyncRequest(message *Message.Message) (string, error) {
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

func (messageHandler *ConcurrentMessageHandler) AddAsyncMessageHandler(topic string, handler func(*Message.Message)) {
	messageHandler.asyncMutex.Lock()
	messageHandler.asyncMessageHandlers[topic] = handler
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *ConcurrentMessageHandler) AddSyncMessageHandler(topic string, handler func(*Message.Message) (string, error)) {
	messageHandler.syncMutex.Lock()
	messageHandler.syncMessageHandlers[topic] = handler
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *ConcurrentMessageHandler) RemoveAsyncMessageHandler(topic string) {
	messageHandler.asyncMutex.Lock()
	delete(messageHandler.asyncMessageHandlers, topic)
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *ConcurrentMessageHandler) RemoveSyncMessageHandler(topic string) {
	messageHandler.syncMutex.Lock()
	delete(messageHandler.syncMessageHandlers, topic)
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *ConcurrentMessageHandler) SetUnknownAsyncHandler(handler func(*Message.Message)) {
	messageHandler.unknwonAsyncTopicHandler = handler
}

func (messageHandler *ConcurrentMessageHandler) SetUnknownSyncHandler(handler func(*Message.Message) (string, error)) {
	messageHandler.unknwonSyncTopicHandler = handler
}

func (messageHandler *ConcurrentMessageHandler) GetAsyncMessageHandler(topic string) func(*Message.Message) {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	return messageHandler.asyncMessageHandlers[topic]
}

func (messageHandler *ConcurrentMessageHandler) GetSyncMessageHandler(topic string) func(*Message.Message) (string, error) {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	return messageHandler.syncMessageHandlers[topic]
}

func (messageHandler *ConcurrentMessageHandler) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}
func (messageHandler *ConcurrentMessageHandler) RetrieveAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *ConcurrentMessageHandler) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}
func (messageHandler *ConcurrentMessageHandler) RetrieveSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *ConcurrentMessageHandler) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}
func (messageHandler *ConcurrentMessageHandler) RetrieveUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
