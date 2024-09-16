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

	SyncRequest(topic string, payload string) (<-chan *Message.Message, error)
	SyncRequestBlocking(topic string, payload string) (*Message.Message, error)
	SyncResponse(message *Message.Message, success bool, payload string) error
	UnprocessedMessagesCount() uint32
}
