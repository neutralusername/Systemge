package SystemgeReceiver

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (receiver *SystemgeReceiver) receive(processingChannel chan func()) {
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Started receiving messages")
	}
	for receiver.processingChannel == processingChannel {
		switch {
		case <-receiver.connection.GetCloseChannel():
			if receiver.infoLogger != nil {
				receiver.infoLogger.Log("Stopped receiving messages")
			}
			receiver.Close()
			return
		default:
			receiver.waitGroup.Add(1)
			messageBytes, err := receiver.connection.ReceiveMessage()
			if err != nil {
				if receiver.warningLogger != nil {
					receiver.warningLogger.Log(Error.New("failed to receive message", err).Error())
				}
				receiver.waitGroup.Done()
				continue
			}
			receiver.messageId++
			messageId := receiver.messageId
			if infoLogger := receiver.infoLogger; infoLogger != nil {
				infoLogger.Log("Received message #" + Helpers.Uint32ToString(messageId))
			}
			func(connection *SystemgeConnection.SystemgeConnection, messageBytes []byte, messageId uint32) {
				processingChannel <- func() {
					err := receiver.processMessage(connection, messageBytes, messageId)
					if err != nil {
						if receiver.warningLogger != nil {
							receiver.warningLogger.Log(Error.New("Failed to process message #"+Helpers.Uint32ToString(messageId), err).Error())
						}
					}
				}
			}(receiver.connection, messageBytes, messageId)
		}
	}

}
