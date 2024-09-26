package SystemgeConnection

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

type ConcurrentMessageHandler struct {
	asyncMessageHandlers AsyncMessageHandlers
	syncMessageHandlers  SyncMessageHandlers
	syncMutex            sync.RWMutex
	asyncMutex           sync.RWMutex

	unknwonAsyncTopicHandler AsyncMessageHandler
	unknwonSyncTopicHandler  SyncMessageHandler

	// metrics
	asyncMessagesHandled  atomic.Uint64
	syncRequestsHandled   atomic.Uint64
	unknownTopicsReceived atomic.Uint64
}

// any number of message handlers may be active at the same time.
func NewConcurrentMessageHandler(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers, unknownTopicAsyncHandler AsyncMessageHandler, unknownTopicSyncHandler SyncMessageHandler) *ConcurrentMessageHandler {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandler := &ConcurrentMessageHandler{
		asyncMessageHandlers:     asyncMessageHandlers,
		syncMessageHandlers:      syncMessageHandlers,
		unknwonAsyncTopicHandler: unknownTopicAsyncHandler,
		unknwonSyncTopicHandler:  unknownTopicSyncHandler,
	}
	return systemgeMessageHandler
}

func (messageHandler *ConcurrentMessageHandler) HandleAsyncMessage(connection SystemgeConnection, message *Message.Message) error {
	messageHandler.asyncMutex.RLock()
	handler, exists := messageHandler.asyncMessageHandlers[message.GetTopic()]
	messageHandler.asyncMutex.RUnlock()
	if !exists {
		if messageHandler.unknwonAsyncTopicHandler != nil {
			messageHandler.asyncMessagesHandled.Add(1)
			messageHandler.unknwonAsyncTopicHandler(connection, message)
			return nil
		} else {
			messageHandler.unknownTopicsReceived.Add(1)
			return errors.New("No handler for async message")
		}
	}
	messageHandler.asyncMessagesHandled.Add(1)
	handler(connection, message)
	return nil
}

func (messageHandler *ConcurrentMessageHandler) HandleSyncRequest(connection SystemgeConnection, message *Message.Message) (string, error) {
	messageHandler.syncMutex.RLock()
	handler, exists := messageHandler.syncMessageHandlers[message.GetTopic()]
	messageHandler.syncMutex.RUnlock()
	if !exists {
		if messageHandler.unknwonSyncTopicHandler != nil {
			messageHandler.syncRequestsHandled.Add(1)
			return messageHandler.unknwonSyncTopicHandler(connection, message)
		} else {
			messageHandler.unknownTopicsReceived.Add(1)
			return "", errors.New("No handler for sync message")
		}
	}
	messageHandler.syncRequestsHandled.Add(1)
	return handler(connection, message)
}

func (messageHandler *ConcurrentMessageHandler) AddAsyncMessageHandler(topic string, handler AsyncMessageHandler) {
	messageHandler.asyncMutex.Lock()
	messageHandler.asyncMessageHandlers[topic] = handler
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *ConcurrentMessageHandler) AddSyncMessageHandler(topic string, handler SyncMessageHandler) {
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

func (messageHandler *ConcurrentMessageHandler) SetUnknownAsyncHandler(handler AsyncMessageHandler) {
	messageHandler.unknwonAsyncTopicHandler = handler
}

func (messageHandler *ConcurrentMessageHandler) SetUnknownSyncHandler(handler SyncMessageHandler) {
	messageHandler.unknwonSyncTopicHandler = handler
}

func (messageHandler *ConcurrentMessageHandler) GetAsyncMessageHandler(topic string) AsyncMessageHandler {
	messageHandler.asyncMutex.RLock()
	defer messageHandler.asyncMutex.RUnlock()
	return messageHandler.asyncMessageHandlers[topic]
}

func (messageHandler *ConcurrentMessageHandler) GetSyncMessageHandler(topic string) SyncMessageHandler {
	messageHandler.syncMutex.RLock()
	defer messageHandler.syncMutex.RUnlock()
	return messageHandler.syncMessageHandlers[topic]
}

func (messageHandler *ConcurrentMessageHandler) GetAsyncTopics() []string {
	messageHandler.asyncMutex.RLock()
	defer messageHandler.asyncMutex.RUnlock()
	topics := make([]string, 0, len(messageHandler.asyncMessageHandlers))
	for topic := range messageHandler.asyncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *ConcurrentMessageHandler) GetSyncTopics() []string {
	messageHandler.syncMutex.RLock()
	defer messageHandler.syncMutex.RUnlock()
	topics := make([]string, 0, len(messageHandler.syncMessageHandlers))
	for topic := range messageHandler.syncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *ConcurrentMessageHandler) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("concurrent_message_handler", Metrics.New(
		map[string]uint64{
			"async_messages_handled":  messageHandler.CheckAsyncMessagesHandled(),
			"sync_requests_handled":   messageHandler.CheckSyncRequestsHandled(),
			"unknown_topics_received": messageHandler.CheckUnknownTopicsReceived(),
		},
	))
	return metricsTypes
}
func (messageHandler *ConcurrentMessageHandler) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("concurrent_message_handler", Metrics.New(
		map[string]uint64{
			"async_messages_handled":  messageHandler.GetAsyncMessagesHandled(),
			"sync_requests_handled":   messageHandler.GetSyncRequestsHandled(),
			"unknown_topics_received": messageHandler.GetUnknownTopicsReceived(),
		},
	))
	return metricsTypes
}

func (messageHandler *ConcurrentMessageHandler) CheckAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}
func (messageHandler *ConcurrentMessageHandler) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *ConcurrentMessageHandler) CheckSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}
func (messageHandler *ConcurrentMessageHandler) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *ConcurrentMessageHandler) CheckUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}
func (messageHandler *ConcurrentMessageHandler) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
