package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Message"
)

type SystemgeConnection interface {
	AbortSyncRequest(syncToken string) error
	AsyncMessage(topic string, payload string) error
	Close() error
	GetAddress() string
	GetAsyncMessagesSent() uint64
	GetByteRateLimiterExceeded() uint64
	GetBytesReceived() uint64
	GetBytesSent() uint64
	GetCloseChannel() <-chan bool
	GetInvalidMessagesReceived() uint64
	GetInvalidSyncResponsesReceived() uint64
	GetMessageRateLimiterExceeded() uint64
	GetMetrics() map[string]uint64
	GetName() string
	GetNextMessage() (*Message.Message, error)
	GetNoSyncResponseReceived() uint64
	GetStatus() int
	GetSyncFailureResponsesReceived() uint64
	GetSyncRequestsSent() uint64
	GetSyncSuccessResponsesReceived() uint64
	GetValidMessagesReceived() uint64
	ProcessMessage(message *Message.Message, messageHandler MessageHandler) error
	RetrieveAsyncMessagesSent() uint64
	RetrieveByteRateLimiterExceeded() uint64
	RetrieveBytesReceived() uint64
	RetrieveBytesSent() uint64
	RetrieveInvalidMessagesReceived() uint64
	RetrieveInvalidSyncResponsesReceived() uint64
	RetrieveMessageRateLimiterExceeded() uint64
	RetrieveMetrics() map[string]uint64
	RetrieveNoSyncResponseReceived() uint64
	RetrieveSyncFailureResponsesReceived() uint64
	RetrieveSyncRequestsSent() uint64
	RetrieveSyncSuccessResponsesReceived() uint64
	RetrieveValidMessagesReceived() uint64
	StartProcessingLoopConcurrently(messageHandler MessageHandler) error
	StartProcessingLoopSequentially(messageHandler MessageHandler) error
	StopProcessingLoop() error
	SyncRequest(topic string, payload string) (<-chan *Message.Message, error)
	SyncRequestBlocking(topic string, payload string) (*Message.Message, error)
	UnprocessedMessagesCount() int64
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
