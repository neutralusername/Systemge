package SystemgeReceiver

import (
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
)

type SystemgeReceiver struct {
	connection     *SystemgeConnection.SystemgeConnection
	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler
	messageChannel chan func()
}

func New(connection *SystemgeConnection.SystemgeConnection, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeReceiver {
	receiver := &SystemgeReceiver{
		connection:     connection,
		messageHandler: messageHandler,
		messageChannel: make(chan func()),
	}
	return receiver
}

func (receiver *SystemgeReceiver) Start() {
	go receiver.receive()
}
