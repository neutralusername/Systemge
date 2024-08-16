package SystemgeReceiver

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
)

func (receiver *SystemgeReceiver) receive(stopChannel chan bool) {
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Started receiving messages")
	}
	for receiver.stopChannel == stopChannel {
		receiver.waitGroup.Add(1)
		messageBytes, err := receiver.connection.ReceiveMessage()
		if err != nil {
			if receiver.errorLogger != nil {
				receiver.errorLogger.Log(Error.New("failed to receive message", err).Error())
			}
			receiver.waitGroup.Done()
			if Tcp.IsConnectionClosed(err) {
				receiver.Stop()
				break
			}
			continue
		}
		receiver.messageId++
		messageId := receiver.messageId
		if infoLogger := receiver.infoLogger; infoLogger != nil {
			infoLogger.Log("Received message #" + Helpers.Uint32ToString(messageId))
		}
		func(connection *SystemgeConnection.SystemgeConnection, messageBytes []byte, messageId uint32) {
			receiver.processingChannel <- func() {
				err := receiver.processMessage(connection, messageBytes, messageId)
				if err != nil {
					if receiver.warningLogger != nil {
						receiver.warningLogger.Log(Error.New("Failed to process message #"+Helpers.Uint32ToString(messageId), err).Error())
					}
				}
			}
		}(receiver.connection, messageBytes, messageId)
	}

	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Stopped receiving messages")
	}
}
