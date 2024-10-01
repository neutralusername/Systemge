package TopicManager

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Metrics"
)

type TopicHandler func(...any) (any, error)
type TopicHandlers map[string]TopicHandler

type Config struct {
	Sequential bool
	Concurrent bool
}

type TopicManager struct {
	topicHandlers       TopicHandlers
	unknownTopicHandler TopicHandler

	mutex sync.RWMutex

	queue             chan *queueStruct
	topicQueues       map[string]chan *queueStruct
	unknownTopicQueue chan *queueStruct

	queueSize       uint32
	topicQueueSize  uint32
	concurrentCalls bool
}

/* type TopicManager interface {
	HandleTopic(string, ...any) (any, error)
	AddTopic(string, TopicHandler)
	RemoveTopic(string)
	GetTopics() []string
	SetUnknownHandler(TopicHandler)
} */

type queueStruct struct {
	topic                string
	args                 []any
	responseAnyChannel   chan any
	responseErrorChannel chan error
}

// modes: (l == large enough to never be full)
// topicQueueSize: 0, queueSize: l concurrentCalls: false -> "sequential"
// topicQueueSize: l, queueSize: l concurrentCalls: false -> "topic exclusive"
// topicQueueSize: 0|l, queueSize: 0|l concurrentCalls: true -> "concurrent"

func NewTopicManager(topicHandlers TopicHandlers, unknownTopicHandler TopicHandler, topicQueueSize uint32, queueSize uint32, concurrentCalls bool) *TopicManager {
	if topicHandlers == nil {
		topicHandlers = make(TopicHandlers)
	}
	topicManager := &TopicManager{
		topicHandlers:       topicHandlers,
		unknownTopicHandler: unknownTopicHandler,
		queue:               make(chan *queueStruct, queueSize),
		topicQueues:         make(map[string]chan *queueStruct),
		unknownTopicQueue:   make(chan *queueStruct, topicQueueSize),
		queueSize:           queueSize,
		topicQueueSize:      topicQueueSize,
		concurrentCalls:     concurrentCalls,
	}
	go topicManager.handleCalls()
	for topic := range topicHandlers {
		topicManager.topicQueues[topic] = make(chan *queueStruct, topicQueueSize)
		go topicManager.handleTopic(topic)
	}
	return topicManager
}

func (topicManager *TopicManager) handleCalls() {
	for {
		queueStruct := <-topicManager.queue
		if queueStruct == nil {
			return
		}
		if queue := topicManager.topicQueues[queueStruct.topic]; queue != nil {
			queue <- queueStruct
		} else if topicManager.unknownTopicQueue != nil {
			topicManager.unknownTopicQueue <- queueStruct
		} else {
			queueStruct.responseAnyChannel <- nil
			queueStruct.responseErrorChannel <- errors.New("no handler for topic")
		}
	}
}

func (topicManager *TopicManager) HandleTopic(topic string, args ...any) (any, error) {
	response := make(chan any)
	err := make(chan error)

	topicManager.queue <- &queueStruct{
		topic:                topic,
		args:                 args,
		responseAnyChannel:   response,
		responseErrorChannel: err,
	}

	return <-response, <-err
}

func (messageHandler *TopicExclusiveMessageHandler) handleMessages() {
	for {
		messageStruct := <-messageHandler.messageQueue
		if messageStruct == nil {
			return
		}
		if messageStruct.syncResponseChannel != nil {
			messageHandler.mutex.RLock()
			handler, exists := messageHandler.syncMessageHandlers[messageStruct.message.GetTopic()]
			messageHandler.mutex.RUnlock()
			if !exists {
				if unknownMessageHandler := messageHandler.unknownSyncTopicHandler; unknownMessageHandler != nil {
					unknownMessageHandler.messageQueue <- messageStruct
				} else {
					messageHandler.unknownTopicsReceived.Add(1)
					messageStruct.syncResponseChannel <- &syncResponseStruct{response: "", err: errors.New("no handler for sync message")}
				}
			} else {
				handler.messageQueue <- messageStruct
			}
		} else {
			messageHandler.asyncMutex.RLock()
			handler, exists := messageHandler.asyncMessageHandlers[messageStruct.message.GetTopic()]
			messageHandler.asyncMutex.RUnlock()
			if !exists {
				if unknownMessageHandler := messageHandler.unknownAsyncTopicHandler; unknownMessageHandler != nil {
					unknownMessageHandler.messageQueue <- messageStruct
				} else {
					messageHandler.unknownTopicsReceived.Add(1)
					messageStruct.asyncErrorChannel <- errors.New("no handler for async message")
				}
			} else {
				handler.messageQueue <- messageStruct
			}
		}
	}
}

func (messageHandler *TopicExclusiveMessageHandler) Close() {
	close(messageHandler.messageQueue)
	messageHandler.asyncMutex.Lock()
	for key, handler := range messageHandler.asyncMessageHandlers {
		close(handler.messageQueue)
		delete(messageHandler.asyncMessageHandlers, key)
	}
	if messageHandler.unknownAsyncTopicHandler != nil {
		close(messageHandler.unknownAsyncTopicHandler.messageQueue)
		messageHandler.unknownAsyncTopicHandler = nil
	}
	messageHandler.asyncMutex.Unlock()

	messageHandler.mutex.Lock()
	for key, handler := range messageHandler.syncMessageHandlers {
		close(handler.messageQueue)
		delete(messageHandler.syncMessageHandlers, key)
	}
	if messageHandler.unknownSyncTopicHandler != nil {
		close(messageHandler.unknownSyncTopicHandler.messageQueue)
		messageHandler.unknownSyncTopicHandler = nil
	}
	messageHandler.mutex.Unlock()
}

func (messageHandler *TopicExclusiveMessageHandler) handleAsyncMessages(asyncMessageHandler *asyncMessageHandler) {
	for {
		messageStruct := <-asyncMessageHandler.messageQueue
		if messageStruct == nil {
			return
		}
		asyncMessageHandler.messageHandler(messageStruct.connection, messageStruct.message)
		messageStruct.asyncErrorChannel <- nil
	}
}

func (messageHandler *TopicExclusiveMessageHandler) handleSyncMessages(syncMessageHandler *syncMessageHandler) {
	for {
		messageStruct := <-syncMessageHandler.messageQueue
		if messageStruct == nil {
			return
		}
		response, err := syncMessageHandler.messageHandler(messageStruct.connection, messageStruct.message)
		messageStruct.syncResponseChannel <- &syncResponseStruct{
			response: response,
			err:      err,
		}
	}
}

func (messageHandler *TopicExclusiveMessageHandler) AddAsyncMessageHandler(topic string, handler AsyncMessageHandler) {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	if messageHandler.asyncMessageHandlers[topic] != nil {
		return
	}
	asyncMessageHandler := &asyncMessageHandler{
		messageHandler: handler,
		messageQueue:   make(chan *queueStruct, messageHandler.queueSize),
	}
	messageHandler.asyncMessageHandlers[topic] = asyncMessageHandler
	go messageHandler.handleAsyncMessages(asyncMessageHandler)
}

func (messageHandler *TopicExclusiveMessageHandler) AddSyncMessageHandler(topic string, handler SyncMessageHandler) {
	messageHandler.mutex.Lock()
	defer messageHandler.mutex.Unlock()
	if messageHandler.syncMessageHandlers[topic] != nil {
		return
	}
	syncMessageHandler := &syncMessageHandler{
		messageHandler: handler,
		messageQueue:   make(chan *queueStruct, messageHandler.queueSize),
	}
	messageHandler.syncMessageHandlers[topic] = syncMessageHandler
	go messageHandler.handleSyncMessages(syncMessageHandler)
}

func (messageHandler *TopicExclusiveMessageHandler) SetUnknownAsyncHandler(handler AsyncMessageHandler) {
	if messageHandler.unknownAsyncTopicHandler != nil {
		close(messageHandler.unknownAsyncTopicHandler.messageQueue)
	}
	if handler != nil {
		messageHandler.unknownAsyncTopicHandler = &asyncMessageHandler{
			messageHandler: handler,
			messageQueue:   make(chan *queueStruct, messageHandler.queueSize),
		}
		go messageHandler.handleAsyncMessages(messageHandler.unknownAsyncTopicHandler)
	} else {
		messageHandler.unknownAsyncTopicHandler = nil
	}
}

func (messageHandler *TopicExclusiveMessageHandler) SetUnknownSyncHandler(handler SyncMessageHandler) {
	if messageHandler.unknownSyncTopicHandler != nil {
		close(messageHandler.unknownSyncTopicHandler.messageQueue)
	}
	if handler != nil {
		messageHandler.unknownSyncTopicHandler = &syncMessageHandler{
			messageHandler: handler,
			messageQueue:   make(chan *queueStruct, messageHandler.queueSize),
		}
		go messageHandler.handleSyncMessages(messageHandler.unknownSyncTopicHandler)
	} else {
		messageHandler.unknownSyncTopicHandler = nil
	}
}

func (messageHandler *TopicExclusiveMessageHandler) RemoveAsyncMessageHandler(topic string) {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	if handler, exists := messageHandler.asyncMessageHandlers[topic]; !exists {
		close(handler.messageQueue)
		delete(messageHandler.asyncMessageHandlers, topic)
	}
}

func (messageHandler *TopicExclusiveMessageHandler) RemoveSyncMessageHandler(topic string) {
	messageHandler.mutex.Lock()
	defer messageHandler.mutex.Unlock()
	if handler, exists := messageHandler.syncMessageHandlers[topic]; !exists {
		close(handler.messageQueue)
		delete(messageHandler.syncMessageHandlers, topic)
	}
}

func (messageHandler *TopicExclusiveMessageHandler) GetAsyncMessageHandler(topic string) AsyncMessageHandler {
	messageHandler.asyncMutex.RLock()
	defer messageHandler.asyncMutex.RUnlock()
	if handler, exists := messageHandler.asyncMessageHandlers[topic]; exists {
		return handler.messageHandler
	}
	return nil
}

func (messageHandler *TopicExclusiveMessageHandler) GetSyncMessageHandler(topic string) SyncMessageHandler {
	messageHandler.mutex.RLock()
	defer messageHandler.mutex.RUnlock()
	if handler, exists := messageHandler.syncMessageHandlers[topic]; exists {
		return handler.messageHandler
	}
	return nil
}

func (messageHandler *TopicExclusiveMessageHandler) GetAsyncTopics() []string {
	messageHandler.asyncMutex.RLock()
	defer messageHandler.asyncMutex.RUnlock()
	topics := make([]string, 0, len(messageHandler.asyncMessageHandlers))
	for topic := range messageHandler.asyncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *TopicExclusiveMessageHandler) GetSyncTopics() []string {
	messageHandler.mutex.RLock()
	defer messageHandler.mutex.RUnlock()
	topics := make([]string, 0, len(messageHandler.syncMessageHandlers))
	for topic := range messageHandler.syncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *TopicExclusiveMessageHandler) CheckMetrics() Metrics.MetricsTypes {
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
func (messageHandler *TopicExclusiveMessageHandler) GetMetrics() Metrics.MetricsTypes {
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

func (messageHandler *TopicExclusiveMessageHandler) CheckAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}
func (messageHandler *TopicExclusiveMessageHandler) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *TopicExclusiveMessageHandler) CheckSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}
func (messageHandler *TopicExclusiveMessageHandler) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *TopicExclusiveMessageHandler) CheckUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}
func (messageHandler *TopicExclusiveMessageHandler) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
