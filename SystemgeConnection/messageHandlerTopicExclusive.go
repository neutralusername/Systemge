package SystemgeConnection

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

type TopicExclusiveMessageHandler struct {
	asyncMessageHandlers map[string]*asyncMessageHandler
	syncMessageHandlers  map[string]*syncMessageHandler

	syncMutex  sync.Mutex
	asyncMutex sync.Mutex

	messageQueue chan *queueStruct

	unknownAsyncTopicHandler *asyncMessageHandler
	unknownSyncTopicHandler  *syncMessageHandler

	queueSize int

	// metrics
	asyncMessagesHandled  atomic.Uint64
	syncRequestsHandled   atomic.Uint64
	unknownTopicsReceived atomic.Uint64
}

type asyncMessageHandler struct {
	messageHandler AsyncMessageHandler
	messageQueue   chan *queueStruct
}

type syncMessageHandler struct {
	messageHandler SyncMessageHandler
	messageQueue   chan *queueStruct
}

// one message handler of each topic may be active at the same time. messages are handled in the order they were added to the queue.
func NewTopicExclusiveMessageHandler(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers, unknownTopicAsyncHandler AsyncMessageHandler, unknownTopicSyncHandler SyncMessageHandler, queueSize int) *TopicExclusiveMessageHandler {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandler := &TopicExclusiveMessageHandler{
		asyncMessageHandlers: make(map[string]*asyncMessageHandler),
		syncMessageHandlers:  make(map[string]*syncMessageHandler),

		messageQueue: make(chan *queueStruct, queueSize),

		queueSize: queueSize,
	}
	if unknownTopicAsyncHandler != nil {
		systemgeMessageHandler.unknownAsyncTopicHandler = &asyncMessageHandler{
			messageHandler: unknownTopicAsyncHandler,
			messageQueue:   make(chan *queueStruct, queueSize),
		}
		go systemgeMessageHandler.handleAsyncMessages(systemgeMessageHandler.unknownAsyncTopicHandler)
	}
	if unknownTopicSyncHandler != nil {
		systemgeMessageHandler.unknownSyncTopicHandler = &syncMessageHandler{
			messageHandler: unknownTopicSyncHandler,
			messageQueue:   make(chan *queueStruct, queueSize),
		}
		go systemgeMessageHandler.handleSyncMessages(systemgeMessageHandler.unknownSyncTopicHandler)
	}
	for topic, handler := range asyncMessageHandlers {
		systemgeMessageHandler.asyncMessageHandlers[topic] = &asyncMessageHandler{
			messageHandler: handler,
			messageQueue:   make(chan *queueStruct, queueSize),
		}
		go systemgeMessageHandler.handleAsyncMessages(systemgeMessageHandler.asyncMessageHandlers[topic])
	}
	for topic, handler := range syncMessageHandlers {
		systemgeMessageHandler.syncMessageHandlers[topic] = &syncMessageHandler{
			messageHandler: handler,
			messageQueue:   make(chan *queueStruct, queueSize),
		}
		go systemgeMessageHandler.handleSyncMessages(systemgeMessageHandler.syncMessageHandlers[topic])
	}
	go systemgeMessageHandler.handleMessages()
	return systemgeMessageHandler
}

func (messageHandler *TopicExclusiveMessageHandler) HandleAsyncMessage(connection SystemgeConnection, message *Message.Message) error {
	response := make(chan error)
	select {
	case messageHandler.messageQueue <- &queueStruct{
		message:           message,
		asyncErrorChannel: response,
		connection:        connection,
	}:
		return <-response
	default:
		return Error.New("Message queue is full", nil)
	}
}

func (messageHandler *TopicExclusiveMessageHandler) HandleSyncRequest(connection SystemgeConnection, message *Message.Message) (string, error) {
	response := make(chan *syncResponseStruct)
	select {
	case messageHandler.messageQueue <- &queueStruct{
		message:             message,
		syncResponseChannel: response,
		connection:          connection,
	}:
		responseStruct := <-response
		return responseStruct.response, responseStruct.err
	default:
		return "", Error.New("Message queue is full", nil)
	}
}

func (messageHandler *TopicExclusiveMessageHandler) handleMessages() {
	for {
		messageStruct := <-messageHandler.messageQueue
		if messageStruct == nil {
			return
		}
		if messageStruct.syncResponseChannel != nil {
			messageHandler.syncMutex.Lock()
			handler, exists := messageHandler.syncMessageHandlers[messageStruct.message.GetTopic()]
			messageHandler.syncMutex.Unlock()
			if !exists {
				if unknownMessageHandler := messageHandler.unknownSyncTopicHandler; unknownMessageHandler != nil {
					unknownMessageHandler.messageQueue <- messageStruct
				} else {
					messageHandler.unknownTopicsReceived.Add(1)
					messageStruct.syncResponseChannel <- &syncResponseStruct{response: "", err: Error.New("No handler for sync message", nil)}
				}
			} else {
				handler.messageQueue <- messageStruct
			}
		} else {
			messageHandler.asyncMutex.Lock()
			handler, exists := messageHandler.asyncMessageHandlers[messageStruct.message.GetTopic()]
			messageHandler.asyncMutex.Unlock()
			if !exists {
				if unknownMessageHandler := messageHandler.unknownAsyncTopicHandler; unknownMessageHandler != nil {
					unknownMessageHandler.messageQueue <- messageStruct
				} else {
					messageHandler.unknownTopicsReceived.Add(1)
					messageStruct.asyncErrorChannel <- Error.New("No handler for async message", nil)
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

	messageHandler.syncMutex.Lock()
	for key, handler := range messageHandler.syncMessageHandlers {
		close(handler.messageQueue)
		delete(messageHandler.syncMessageHandlers, key)
	}
	if messageHandler.unknownSyncTopicHandler != nil {
		close(messageHandler.unknownSyncTopicHandler.messageQueue)
		messageHandler.unknownSyncTopicHandler = nil
	}
	messageHandler.syncMutex.Unlock()
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
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
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
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	if handler, exists := messageHandler.syncMessageHandlers[topic]; !exists {
		close(handler.messageQueue)
		delete(messageHandler.syncMessageHandlers, topic)
	}
}

func (messageHandler *TopicExclusiveMessageHandler) GetAsyncMessageHandler(topic string) AsyncMessageHandler {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	if handler, exists := messageHandler.asyncMessageHandlers[topic]; exists {
		return handler.messageHandler
	}
	return nil
}

func (messageHandler *TopicExclusiveMessageHandler) GetSyncMessageHandler(topic string) SyncMessageHandler {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	if handler, exists := messageHandler.syncMessageHandlers[topic]; exists {
		return handler.messageHandler
	}
	return nil
}

func (messageHandler *TopicExclusiveMessageHandler) GetAsyncTopics() []string {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	topics := make([]string, 0, len(messageHandler.asyncMessageHandlers))
	for topic := range messageHandler.asyncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *TopicExclusiveMessageHandler) GetSyncTopics() []string {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
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
