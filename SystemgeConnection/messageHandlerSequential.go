package SystemgeConnection

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

type SequentialMessageHandler struct {
	asyncMessageHandlers AsyncMessageHandlers
	syncMessageHandlers  SyncMessageHandlers

	syncMutex  sync.Mutex
	asyncMutex sync.Mutex

	messageQueue chan *queueStruct

	unknwonAsyncTopicHandler AsyncMessageHandler
	unknwonSyncTopicHandler  SyncMessageHandler

	// metrics
	asyncMessagesHandled  atomic.Uint64
	syncRequestsHandled   atomic.Uint64
	unknownTopicsReceived atomic.Uint64
}

type queueStruct struct {
	message             *Message.Message
	syncResponseChannel chan *syncResponseStruct
	asyncErrorChannel   chan error
	connection          SystemgeConnection
}

type syncResponseStruct struct {
	response string
	err      error
}

// one message handler can be active at the same time.
// requires a call to Close() to stop the message handler (otherwise it will keep running until the program ends).
// Handle calls after Close() will cause a panic.
func NewSequentialMessageHandler(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers, unknownTopicAsyncHandler AsyncMessageHandler, unknownTopicSyncHandler SyncMessageHandler, queueSize int) *SequentialMessageHandler {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandler := &SequentialMessageHandler{
		asyncMessageHandlers:     asyncMessageHandlers,
		syncMessageHandlers:      syncMessageHandlers,
		unknwonAsyncTopicHandler: unknownTopicAsyncHandler,
		unknwonSyncTopicHandler:  unknownTopicSyncHandler,
		messageQueue:             make(chan *queueStruct, queueSize),
	}
	go systemgeMessageHandler.handleMessages()
	return systemgeMessageHandler
}

func (messageHandler *SequentialMessageHandler) Close() {
	close(messageHandler.messageQueue)
}

func (messageHandler *SequentialMessageHandler) handleMessages() {
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
				if unknownMessageHandler := messageHandler.unknwonSyncTopicHandler; unknownMessageHandler != nil {
					messageHandler.syncRequestsHandled.Add(1)
					response, err := unknownMessageHandler(messageStruct.connection, messageStruct.message)
					messageStruct.syncResponseChannel <- &syncResponseStruct{response: response, err: err}
				} else {
					messageHandler.unknownTopicsReceived.Add(1)
					messageStruct.syncResponseChannel <- &syncResponseStruct{response: "", err: Error.New("No handler for sync message", nil)}
				}
			} else {
				messageHandler.syncRequestsHandled.Add(1)
				response, err := handler(messageStruct.connection, messageStruct.message)
				messageStruct.syncResponseChannel <- &syncResponseStruct{response: response, err: err}
			}
		} else {
			messageHandler.asyncMutex.Lock()
			handler, exists := messageHandler.asyncMessageHandlers[messageStruct.message.GetTopic()]
			messageHandler.asyncMutex.Unlock()
			if !exists {
				if unknownMessageHandler := messageHandler.unknwonAsyncTopicHandler; unknownMessageHandler != nil {
					messageHandler.asyncMessagesHandled.Add(1)
					unknownMessageHandler(messageStruct.connection, messageStruct.message)
					messageStruct.asyncErrorChannel <- nil
				} else {
					messageHandler.unknownTopicsReceived.Add(1)
					messageStruct.asyncErrorChannel <- Error.New("No handler for async message", nil)
				}
			} else {
				messageHandler.asyncMessagesHandled.Add(1)
				handler(messageStruct.connection, messageStruct.message)
				messageStruct.asyncErrorChannel <- nil
			}
		}
	}
}

func (messageHandler *SequentialMessageHandler) HandleAsyncMessage(connection SystemgeConnection, message *Message.Message) error {
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

func (messageHandler *SequentialMessageHandler) HandleSyncRequest(connection SystemgeConnection, message *Message.Message) (string, error) {
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

func (messageHandler *SequentialMessageHandler) AddAsyncMessageHandler(topic string, handler AsyncMessageHandler) {
	messageHandler.asyncMutex.Lock()
	messageHandler.asyncMessageHandlers[topic] = handler
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *SequentialMessageHandler) AddSyncMessageHandler(topic string, handler SyncMessageHandler) {
	messageHandler.syncMutex.Lock()
	messageHandler.syncMessageHandlers[topic] = handler
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *SequentialMessageHandler) RemoveAsyncMessageHandler(topic string) {
	messageHandler.asyncMutex.Lock()
	delete(messageHandler.asyncMessageHandlers, topic)
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *SequentialMessageHandler) RemoveSyncMessageHandler(topic string) {
	messageHandler.syncMutex.Lock()
	delete(messageHandler.syncMessageHandlers, topic)
	messageHandler.syncMutex.Unlock()
}

func (messageHandler *SequentialMessageHandler) GetAsyncTopics() []string {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	topics := make([]string, 0, len(messageHandler.asyncMessageHandlers))
	for topic := range messageHandler.asyncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *SequentialMessageHandler) GetSyncTopics() []string {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	topics := make([]string, 0, len(messageHandler.syncMessageHandlers))
	for topic := range messageHandler.syncMessageHandlers {
		topics = append(topics, topic)
	}
	return topics
}

func (messageHandler *SequentialMessageHandler) SetUnknownAsyncHandler(handler AsyncMessageHandler) {
	messageHandler.unknwonAsyncTopicHandler = handler
}

func (messageHandler *SequentialMessageHandler) SetUnknownSyncHandler(handler SyncMessageHandler) {
	messageHandler.unknwonSyncTopicHandler = handler
}

func (messageHandler *SequentialMessageHandler) GetAsyncMessageHandler(topic string) AsyncMessageHandler {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	return messageHandler.asyncMessageHandlers[topic]
}

func (messageHandler *SequentialMessageHandler) GetSyncMessageHandler(topic string) SyncMessageHandler {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	return messageHandler.syncMessageHandlers[topic]
}

func (messageHandler *SequentialMessageHandler) CheckMetrics() Metrics.MetricsTypes {
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
func (messageHandler *SequentialMessageHandler) GetMetrics() Metrics.MetricsTypes {
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

func (messageHandler *SequentialMessageHandler) CheckAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}
func (messageHandler *SequentialMessageHandler) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *SequentialMessageHandler) CheckSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}
func (messageHandler *SequentialMessageHandler) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *SequentialMessageHandler) CheckUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}
func (messageHandler *SequentialMessageHandler) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
