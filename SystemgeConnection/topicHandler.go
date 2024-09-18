package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

type TopicHandler interface {
	HandleMessage(connection SystemgeConnection, message *Message.Message) error
	AddMessageHandlerFunc(topic string, handler MessageHandlerFunc)
	RemoveMessageHandlerFunc(topic string)
	GetMessageHandlerFunc(topic string) MessageHandlerFunc
	GetTopics() []string

	CheckMetrics() Metrics.MetricsTypes
	GetMetrics() Metrics.MetricsTypes

	CheckMessagesHandled() uint64
	GetMessagesHandled() uint64

	CheckUnknownTopicsReceived() uint64
	GetUnknownTopicsReceived() uint64
}
