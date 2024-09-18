package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

// handles attributes of message in provided order
type AttributeHandler interface {
	HandleMessage(connection SystemgeConnection, message *Message.Message) error
	AddMessageHandlerFunc(attribute string, handler MessageHandlerFunc, abortOnError bool)
	RemoveMessageHandlerFunc(attribute string)
	GetMessageHandlerFunc(attribute string) MessageHandlerFunc
	GetAttributes() []string

	CheckMetrics() Metrics.MetricsTypes
	GetMetrics() Metrics.MetricsTypes

	CheckMessagesHandled() uint64
	GetMessagesHandled() uint64

	CheckAttributesHandled() uint64
	GetAttributesHandled() uint64

	CheckUnknownAttributesReceived() uint64
	GetUnknownAttributesReceived() uint64
}

type Attribute struct {
	abortOnError  bool
	attributeCall func(SystemgeConnection, *Message.Message) error
}

/*
func (connection SystemgeConnection) ExecuteAttribute(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; !ok {
		return Error.New("No attribute found", nil)
	}
	return connection.attributes[attribute].attributeCall()
}

func (connection *TcpSystemgeConnection) NewAttribute(attribute string, abortOnError bool) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		return Error.New("Attribute already exists", nil)
	}
	connection.attributes[attribute] = &Attribute{
		abortOnError:  abortOnError,
		attributeCall: func() error { return nil },
	}
	return nil
}

func (connection *TcpSystemgeConnection) RemoveAttribute(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; !ok {
		return Error.New("No attribute found", nil)
	}
	delete(connection.attributes, attribute)
	return nil
}

func (connection *TcpSystemgeConnection) GetAttribute(attribute string) (*Attribute, error) {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if attribute, ok := connection.attributes[attribute]; ok {
		return attribute, nil
	}
	return nil, Error.New("No attribute found", nil)
}
*/
