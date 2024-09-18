package SystemgeConnection

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

type TopicHandlerConcurrently struct {
	messageHandlerFuncs map[string]MessageHandlerFunc
	mutex               sync.Mutex

	unknownTopicHandlerFunc MessageHandlerFunc

	// metrics
	messagesHandled       atomic.Uint64
	unknownTopicsReceived atomic.Uint64
}

// any number of message handlers may be active at the same time.
func NewConcurrentMessageHandler(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers, unknownTopicAsyncHandler AsyncMessageHandler, unknownTopicSyncHandler SyncMessageHandler) *TopicHandlerConcurrently {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandler := &TopicHandlerConcurrently{
		asyncMessageHandlers:     asyncMessageHandlers,
		syncMessageHandlers:      syncMessageHandlers,
		unknwonAsyncTopicHandler: unknownTopicAsyncHandler,
		unknwonSyncTopicHandler:  unknownTopicSyncHandler,
	}
	return systemgeMessageHandler
}

func (messageHandler *TopicHandlerConcurrently) HandleAsyncMessage(connection SystemgeConnection, message *Message.Message) error {
	messageHandler.asyncMutex.Lock()
	handler, exists := messageHandler.asyncMessageHandlers[message.GetTopic()]
	messageHandler.asyncMutex.Unlock()
	if !exists {
		if messageHandler.unknwonAsyncTopicHandler != nil {
			messageHandler.asyncMessagesHandled.Add(1)
			messageHandler.unknwonAsyncTopicHandler(connection, message)
			return nil
		} else {
			messageHandler.unknownTopicsReceived.Add(1)
			return Error.New("No handler for async message", nil)
		}
	}
	messageHandler.asyncMessagesHandled.Add(1)
	handler(connection, message)
	return nil
}

func (messageHandler *TopicHandlerConcurrently) HandleSyncRequest(connection SystemgeConnection, message *Message.Message) (string, error) {
	messageHandler.syncMutex.Lock()
	handler, exists := messageHandler.syncMessageHandlers[message.GetTopic()]
	messageHandler.syncMutex.Unlock()
	if !exists {
		if messageHandler.unknwonSyncTopicHandler != nil {
			messageHandler.syncRequestsHandled.Add(1)
			return messageHandler.unknwonSyncTopicHandler(connection, message)
		} else {
			messageHandler.unknownTopicsReceived.Add(1)
			return "", Error.New("No handler for sync message", nil)
		}
	}
	messageHandler.syncRequestsHandled.Add(1)
	return handler(connection, message)
}

func (messageHandler *TopicHandlerConcurrently) AddAsyncMessageHandler(topic string, handler AsyncMessageHandler) {
	messageHandler.asyncMutex.Lock()
	messageHandler.asyncMessageHandlers[topic] = handler
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *TopicHandlerConcurrently) AddSyncMessageHandler(topic string, handler SyncMessageHandler) {
	messageHandler.syncMutex.Lock()
	messageHandler.syncMessageHandlers[topic] = handler
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *TopicHandlerConcurrently) RemoveAsyncMessageHandler(topic string) {
	messageHandler.asyncMutex.Lock()
	delete(messageHandler.asyncMessageHandlers, topic)
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *TopicHandlerConcurrently) RemoveSyncMessageHandler(topic string) {
	messageHandler.syncMutex.Lock()
	delete(messageHandler.syncMessageHandlers, topic)
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *TopicHandlerConcurrently) SetUnknownAsyncHandler(handler AsyncMessageHandler) {
	messageHandler.unknwonAsyncTopicHandler = handler
}

func (messageHandler *TopicHandlerConcurrently) SetUnknownSyncHandler(handler SyncMessageHandler) {
	messageHandler.unknwonSyncTopicHandler = handler
}

func (messageHandler *TopicHandlerConcurrently) GetAsyncMessageHandler(topic string) AsyncMessageHandler {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	return messageHandler.asyncMessageHandlers[topic]
}

func (messageHandler *TopicHandlerConcurrently) GetSyncMessageHandler(topic string) SyncMessageHandler {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	return messageHandler.syncMessageHandlers[topic]
}

func (messageHandler *TopicHandlerConcurrently) GetAsyncTopics() []string {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	topics := make([]string, 0, len(messageHandler.asyncMessageHandlers))
	for topic := range messageHandler.asyncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *TopicHandlerConcurrently) GetSyncTopics() []string {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	topics := make([]string, 0, len(messageHandler.syncMessageHandlers))
	for topic := range messageHandler.syncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *TopicHandlerConcurrently) CheckMetrics() Metrics.MetricsTypes {
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
func (messageHandler *TopicHandlerConcurrently) GetMetrics() Metrics.MetricsTypes {
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

func (messageHandler *TopicHandlerConcurrently) CheckAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}
func (messageHandler *TopicHandlerConcurrently) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *TopicHandlerConcurrently) CheckSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}
func (messageHandler *TopicHandlerConcurrently) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *TopicHandlerConcurrently) CheckUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}
func (messageHandler *TopicHandlerConcurrently) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
