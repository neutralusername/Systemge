package SystemgeMessageHandler

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

type SequentialMessageHandler struct {
	asyncMessageHandlers AsyncMessageHandlers
	syncMessageHandlers  SyncMessageHandlers

	syncMutex  sync.Mutex
	asyncMutex sync.Mutex

	messageQueue chan *queueStruct

	unknwonAsyncTopicHandler func(*Message.Message)
	unknwonSyncTopicHandler  func(*Message.Message) (string, error)

	// metrics
	asyncMessagesHandled  atomic.Uint64
	syncRequestsHandled   atomic.Uint64
	unknownTopicsReceived atomic.Uint64
}

type queueStruct struct {
	message         *Message.Message
	responseChannel chan *syncResponseStruct
}

type syncResponseStruct struct {
	response string
	err      error
}

// requires a call to Close() to stop the message handler (otherwise it will keep running until the program ends)
func NewSequentialMessageHandler(asyncMessageHandlers AsyncMessageHandlers, syncMessageHandlers SyncMessageHandlers, unknownTopicAsyncHandler *AsyncMessageHandler, unknownTopicSyncHandler *SyncMessageHandler, queueSize int) *SequentialMessageHandler {
	if asyncMessageHandlers == nil {
		asyncMessageHandlers = make(AsyncMessageHandlers)
	}
	if syncMessageHandlers == nil {
		syncMessageHandlers = make(SyncMessageHandlers)
	}
	systemgeMessageHandler := &SequentialMessageHandler{
		asyncMessageHandlers:     asyncMessageHandlers,
		syncMessageHandlers:      syncMessageHandlers,
		unknwonAsyncTopicHandler: *unknownTopicAsyncHandler,
		unknwonSyncTopicHandler:  *unknownTopicSyncHandler,
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
		if messageStruct.responseChannel != nil {
			messageHandler.syncMutex.Lock()
			handler, exists := messageHandler.syncMessageHandlers[messageStruct.message.GetTopic()]
			messageHandler.syncMutex.Unlock()
			if !exists {
				messageHandler.unknownTopicsReceived.Add(1)
				if messageHandler.unknwonSyncTopicHandler != nil {
					messageHandler.syncRequestsHandled.Add(1)
					response, err := messageHandler.unknwonSyncTopicHandler(messageStruct.message)
					messageStruct.responseChannel <- &syncResponseStruct{response: response, err: err}
					continue
				} else {
					messageStruct.responseChannel <- &syncResponseStruct{response: "", err: Error.New("No handler for sync message", nil)}
					continue
				}
			}
			messageHandler.syncRequestsHandled.Add(1)
			response, err := handler(messageStruct.message)
			messageStruct.responseChannel <- &syncResponseStruct{response: response, err: err}
		} else {
			messageHandler.asyncMutex.Lock()
			handler, exists := messageHandler.asyncMessageHandlers[messageStruct.message.GetTopic()]
			messageHandler.asyncMutex.Unlock()
			if !exists {
				messageHandler.unknownTopicsReceived.Add(1)
				if messageHandler.unknwonAsyncTopicHandler != nil {
					messageHandler.asyncMessagesHandled.Add(1)
					messageHandler.unknwonAsyncTopicHandler(messageStruct.message)
				}
				continue
			}
			messageHandler.asyncMessagesHandled.Add(1)
			handler(messageStruct.message)
		}
	}
}

func (messageHandler *SequentialMessageHandler) HandleAsyncMessage(message *Message.Message) error {
	if len(messageHandler.messageQueue) == cap(messageHandler.messageQueue) {
		return Error.New("Message queue is full", nil)
	}
	select {
	case messageHandler.messageQueue <- &queueStruct{message: message, responseChannel: nil}:
		return nil
	default:
		return Error.New("Message queue is full", nil)
	}
}

func (messageHandler *SequentialMessageHandler) HandleSyncRequest(message *Message.Message) (string, error) {
	response := make(chan *syncResponseStruct)
	select {
	case messageHandler.messageQueue <- &queueStruct{message: message, responseChannel: response}:
		responseStruct := <-response
		return responseStruct.response, responseStruct.err
	default:
		return "", Error.New("Message queue is full", nil)
	}
}

func (messageHandler *SequentialMessageHandler) AddAsyncMessageHandler(topic string, handler func(*Message.Message)) {
	messageHandler.asyncMutex.Lock()
	messageHandler.asyncMessageHandlers[topic] = handler
	messageHandler.asyncMutex.Unlock()
}

func (messageHandler *SequentialMessageHandler) AddSyncMessageHandler(topic string, handler func(*Message.Message) (string, error)) {
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

func (messageHandler *SequentialMessageHandler) SetUnknownAsyncHandler(handler func(*Message.Message)) {
	messageHandler.unknwonAsyncTopicHandler = handler
}

func (messageHandler *SequentialMessageHandler) SetUnknownSyncHandler(handler func(*Message.Message) (string, error)) {
	messageHandler.unknwonSyncTopicHandler = handler
}

func (messageHandler *SequentialMessageHandler) GetAsyncMessageHandler(topic string) func(*Message.Message) {
	messageHandler.asyncMutex.Lock()
	defer messageHandler.asyncMutex.Unlock()
	return messageHandler.asyncMessageHandlers[topic]
}

func (messageHandler *SequentialMessageHandler) GetSyncMessageHandler(topic string) func(*Message.Message) (string, error) {
	messageHandler.syncMutex.Lock()
	defer messageHandler.syncMutex.Unlock()
	return messageHandler.syncMessageHandlers[topic]
}

func (messageHandler *SequentialMessageHandler) GetAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Load()
}
func (messageHandler *SequentialMessageHandler) RetrieveAsyncMessagesHandled() uint64 {
	return messageHandler.asyncMessagesHandled.Swap(0)
}

func (messageHandler *SequentialMessageHandler) GetSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Load()
}
func (messageHandler *SequentialMessageHandler) RetrieveSyncRequestsHandled() uint64 {
	return messageHandler.syncRequestsHandled.Swap(0)
}

func (messageHandler *SequentialMessageHandler) GetUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Load()
}
func (messageHandler *SequentialMessageHandler) RetrieveUnknownTopicsReceived() uint64 {
	return messageHandler.unknownTopicsReceived.Swap(0)
}
