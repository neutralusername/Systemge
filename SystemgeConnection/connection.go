package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

type SystemgeConnection interface {
	AbortSyncRequest(syncToken string) error
	AsyncMessage(topic string, payload string) error
	Close() error
	GetAddress() string
	GetCloseChannel() <-chan bool
	GetDefaultCommands() Commands.Handlers
	GetName() string
	GetNextMessage() (*Message.Message, error)
	GetStatus() int
	CheckInvalidMessagesReceived() uint64
	CheckInvalidSyncResponsesReceived() uint64
	CheckMessageRateLimiterExceeded() uint64
	CheckMetrics() map[string]*Metrics.Metrics
	CheckNoSyncResponseReceived() uint64
	CheckSyncFailureResponsesReceived() uint64
	CheckSyncRequestsSent() uint64
	CheckSyncSuccessResponsesReceived() uint64
	CheckValidMessagesReceived() uint64
	CheckAsyncMessagesSent() uint64
	CheckByteRateLimiterExceeded() uint64
	CheckBytesReceived() uint64
	CheckBytesSent() uint64
	IsProcessingLoopRunning() bool
	ProcessMessage(message *Message.Message, messageHandler MessageHandler) error
	GetAsyncMessagesSent() uint64
	GetByteRateLimiterExceeded() uint64
	GetBytesReceived() uint64
	GetBytesSent() uint64
	GetInvalidMessagesReceived() uint64
	GetInvalidSyncResponsesReceived() uint64
	GetMessageRateLimiterExceeded() uint64
	GetMetrics() map[string]*Metrics.Metrics
	GetNoSyncResponseReceived() uint64
	GetSyncFailureResponsesReceived() uint64
	GetSyncRequestsSent() uint64
	GetSyncSuccessResponsesReceived() uint64
	GetValidMessagesReceived() uint64
	StartProcessingLoopConcurrently(messageHandler MessageHandler) error
	StartProcessingLoopSequentially(messageHandler MessageHandler) error
	StopProcessingLoop() error
	SyncRequest(topic string, payload string) (<-chan *Message.Message, error)
	SyncRequestBlocking(topic string, payload string) (*Message.Message, error)
	SyncResponse(message *Message.Message, success bool, payload string) error
	UnprocessedMessagesCount() uint32
}

type AsyncMessageHandler func(SystemgeConnection, *Message.Message)
type AsyncMessageHandlers map[string]AsyncMessageHandler

type SyncMessageHandler func(SystemgeConnection, *Message.Message) (string, error)
type SyncMessageHandlers map[string]SyncMessageHandler

func NewAsyncMessageHandlers() AsyncMessageHandlers {
	return make(AsyncMessageHandlers)
}

func (map1 AsyncMessageHandlers) Merge(map2 AsyncMessageHandlers) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 AsyncMessageHandlers) Add(key string, value AsyncMessageHandler) {
	map1[key] = value
}

func (map1 AsyncMessageHandlers) Remove(key string) {
	delete(map1, key)
}

func (map1 AsyncMessageHandlers) Get(key string) (AsyncMessageHandler, bool) {
	value, ok := map1[key]
	return value, ok
}

func NewSyncMessageHandlers() SyncMessageHandlers {
	return make(SyncMessageHandlers)
}

func (map1 SyncMessageHandlers) Merge(map2 SyncMessageHandlers) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 SyncMessageHandlers) Add(key string, value SyncMessageHandler) {
	map1[key] = value
}

func (map1 SyncMessageHandlers) Remove(key string) {
	delete(map1, key)
}

func (map1 SyncMessageHandlers) Get(key string) (SyncMessageHandler, bool) {
	value, ok := map1[key]
	return value, ok
}
