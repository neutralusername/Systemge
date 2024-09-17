package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

type SystemgeConnection interface {
	GetName() string
	GetAddress() string
	GetStatus() int
	Close() error
	GetCloseChannel() <-chan bool
	GetNextMessage() (*Message.Message, error)
	AvailableMessageCount() uint32

	RegisterMessageHandlingLoop(chan<- bool) error
	IsMessageHandlingLoopRegistered() bool
	StopMessageHandlingLoop() error

	AsyncMessage(topic string, payload string) error
	SyncRequest(topic string, payload string) (<-chan *Message.Message, error)
	SyncRequestBlocking(topic string, payload string) (*Message.Message, error)
	AbortSyncRequest(syncToken string) error
	SyncResponse(message *Message.Message, success bool, payload string) error

	GetDefaultCommands() Commands.Handlers

	CheckInvalidMessagesReceived() uint64
	CheckInvalidSyncResponsesReceived() uint64
	CheckMessageRateLimiterExceeded() uint64
	CheckMetrics() Metrics.MetricsTypes
	CheckNoSyncResponseReceived() uint64
	CheckSyncFailureResponsesReceived() uint64
	CheckSyncRequestsSent() uint64
	CheckSyncSuccessResponsesReceived() uint64
	CheckValidMessagesReceived() uint64
	CheckAsyncMessagesSent() uint64
	CheckSyncResponsesSent() uint64
	CheckByteRateLimiterExceeded() uint64
	CheckBytesReceived() uint64
	CheckBytesSent() uint64

	GetAsyncMessagesSent() uint64
	GetSyncResponsesSent() uint64
	GetByteRateLimiterExceeded() uint64
	GetBytesReceived() uint64
	GetBytesSent() uint64
	GetInvalidMessagesReceived() uint64
	GetInvalidSyncResponsesReceived() uint64
	GetMessageRateLimiterExceeded() uint64
	GetMetrics() Metrics.MetricsTypes
	GetNoSyncResponseReceived() uint64
	GetSyncFailureResponsesReceived() uint64
	GetSyncRequestsSent() uint64
	GetSyncSuccessResponsesReceived() uint64
	GetValidMessagesReceived() uint64
}
