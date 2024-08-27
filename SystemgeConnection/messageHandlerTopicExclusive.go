package SystemgeConnection

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

type TopicExclusiveMessageHandler struct {
	asyncMessageHandlers map[string]*asyncMessageHandler
	syncMessageHandlers  map[string]*syncMessageHandler

	syncMutex  sync.Mutex
	asyncMutex sync.Mutex

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

// one message handler of each topic may be active at the same time.
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
	return systemgeMessageHandler
}

func (messageHandler *TopicExclusiveMessageHandler) HandleAsyncMessage(connection *SystemgeConnection, message *Message.Message) error {
	messageHandler.asyncMutex.Lock()
	handler, exists := messageHandler.asyncMessageHandlers[message.GetTopic()]
	messageHandler.asyncMutex.Unlock()
	if !exists {
		if messageHandler.unknownAsyncTopicHandler != nil {
			messageHandler.asyncMessagesHandled.Add(1)
			response := make(chan error)
			select {
			case messageHandler.unknownAsyncTopicHandler.messageQueue <- &queueStruct{
				connection:          connection,
				message:             message,
				syncResponseChannel: nil,
				asyncErrorChannel:   response,
			}:
				return <-response
			default:
				return Error.New("Message queue is full", nil)
			}
		} else {
			messageHandler.unknownTopicsReceived.Add(1)
			return Error.New("No handler for async message", nil)
		}
	}
	messageHandler.asyncMessagesHandled.Add(1)
	response := make(chan error)
	select {
	case handler.messageQueue <- &queueStruct{
		connection:          connection,
		message:             message,
		syncResponseChannel: nil,
		asyncErrorChannel:   response,
	}:
		return <-response
	default:
		return Error.New("Message queue is full", nil)
	}
}

func (messageHandler *TopicExclusiveMessageHandler) HandleSyncRequest(connection *SystemgeConnection, message *Message.Message) (string, error) {
	messageHandler.syncMutex.Lock()
	handler, exists := messageHandler.syncMessageHandlers[message.GetTopic()]
	messageHandler.syncMutex.Unlock()
	if !exists {
		if messageHandler.unknownSyncTopicHandler != nil {
			messageHandler.syncRequestsHandled.Add(1)
			response := make(chan *syncResponseStruct)
			select {
			case messageHandler.unknownSyncTopicHandler.messageQueue <- &queueStruct{
				connection:          connection,
				message:             message,
				syncResponseChannel: response,
			}:
				responseStruct := <-response
				return responseStruct.response, responseStruct.err
			default:
				return "", Error.New("Message queue is full", nil)
			}
		} else {
			messageHandler.unknownTopicsReceived.Add(1)
			return "", Error.New("No handler for sync message", nil)
		}
	}
	messageHandler.syncRequestsHandled.Add(1)
	response := make(chan *syncResponseStruct)
	select {
	case handler.messageQueue <- &queueStruct{
		connection:          connection,
		message:             message,
		syncResponseChannel: response,
	}:
		responseStruct := <-response
		return responseStruct.response, responseStruct.err
	default:
		return "", Error.New("Message queue is full", nil)
	}
}

func (messageHandler *TopicExclusiveMessageHandler) Close() {
	messageHandler.asyncMutex.Lock()
	for key, handler := range messageHandler.asyncMessageHandlers {
		close(handler.messageQueue)
		delete(messageHandler.asyncMessageHandlers, key)
	}
	close(messageHandler.unknownAsyncTopicHandler.messageQueue)
	messageHandler.unknownAsyncTopicHandler = nil
	messageHandler.asyncMutex.Unlock()

	messageHandler.syncMutex.Lock()
	for key, handler := range messageHandler.syncMessageHandlers {
		close(handler.messageQueue)
		delete(messageHandler.syncMessageHandlers, key)
	}
	close(messageHandler.unknownSyncTopicHandler.messageQueue)
	messageHandler.unknownSyncTopicHandler = nil
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

func (messageHandler *TopicExclusiveMessageHandler) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}
func (messageHandler *TopicExclusiveMessageHandler) RetrieveAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *TopicExclusiveMessageHandler) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}
func (messageHandler *TopicExclusiveMessageHandler) RetrieveSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *TopicExclusiveMessageHandler) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}
func (messageHandler *TopicExclusiveMessageHandler) RetrieveUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
