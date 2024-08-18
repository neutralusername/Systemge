package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

func (receiver *SystemgeReceiver) receive(processingChannel chan func()) {
	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Started receiving messages")
	}
	for receiver.processingChannel == processingChannel {
		receiver.waitGroup.Add(1)
		messageBytes, err := receiver.connection.ReceiveMessage()
		if err != nil {
			if receiver.warningLogger != nil {
				receiver.warningLogger.Log(Error.New("failed to receive message", err).Error())
			}
			receiver.waitGroup.Done()
			if Tcp.IsConnectionClosed(err) {
				receiver.connection.Close()
				if receiver.errorLogger != nil {
					receiver.errorLogger.Log("Connection closed")
				}
				if receiver.mailer != nil {
					err := receiver.mailer.Send(Tools.NewMail(nil, "error", Error.New("connection closed", err).Error()))
					if err != nil {
						if receiver.errorLogger != nil {
							receiver.errorLogger.Log(Error.New("failed sending mail", err).Error())
						}
					}
				}
				break
			}
			continue
		}
		receiver.messageId++
		messageId := receiver.messageId
		if infoLogger := receiver.infoLogger; infoLogger != nil {
			infoLogger.Log("Received message #" + Helpers.Uint32ToString(messageId))
		}
		func(connection *SystemgeConnection, messageBytes []byte, messageId uint32) {
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

	if receiver.infoLogger != nil {
		receiver.infoLogger.Log("Stopped receiving messages")
	}
}
