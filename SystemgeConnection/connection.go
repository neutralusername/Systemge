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
	RetrieveNextMessage() (*Message.Message, error)
	AvailableMessageCount() uint32

	StartMessageHandlingLoop(MessageHandler, bool) error
	IsMessageHandlingLoopStarted() bool
	StopMessageHandlingLoop() error
	HandleMessage(message *Message.Message, messageHandler MessageHandler) error

	AsyncMessage(topic string, payload string) error
	SyncRequest(topic string, payload string) (<-chan *Message.Message, error)
	SyncRequestBlocking(topic string, payload string) (*Message.Message, error)
	AbortSyncRequest(syncToken string) error
	SyncResponse(message *Message.Message, success bool, payload string) error

	GetDefaultCommands() Commands.Handlers

	CheckBytesSent() uint64
	CheckBytesReceived() uint64
	CheckAsyncMessagesSent() uint64
	CheckSyncRequestsSent() uint64
	CheckSyncResponsesSent() uint64
	CheckMessagesReceived() uint64
	CheckInvalidMessagesReceived() uint64
	CheckRejectedMessages() uint64
	CheckInvalidSyncResponsesReceived() uint64
	CheckNoSyncResponseReceived() uint64
	CheckSyncFailureResponsesReceived() uint64
	CheckSyncSuccessResponsesReceived() uint64

	GetBytesSent() uint64
	GetBytesReceived() uint64
	GetAsyncMessagesSent() uint64
	GetSyncRequestsSent() uint64
	GetSyncResponsesSent() uint64
	GetMessagesReceived() uint64
	GetInvalidMessagesReceived() uint64
	GetRejectedMessages() uint64
	GetInvalidSyncResponsesReceived() uint64
	GetNoSyncResponseReceived() uint64
	GetSyncFailureResponsesReceived() uint64
	GetSyncSuccessResponsesReceived() uint64

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
