package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpSystemgeConnection) receiveMessageLoop() {
	if connection.infoLogger != nil {
		connection.infoLogger.Log("Started receiving messages")
	}
	for {
		select {
		case <-connection.closeChannel:
			if connection.infoLogger != nil {
				connection.infoLogger.Log("Stopped receiving messages")
			}
			close(connection.receiveLoopStopChannel)
			return
		default:
			select {
			case <-connection.closeChannel:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Stopped receiving messages")
				}
				close(connection.receiveLoopStopChannel)
				return
			case <-connection.messageChannelSemaphore.GetChannel():
				messageBytes, err := connection.messageReceiver.ReceiveNextMessage()
				if err != nil {
					if Tcp.IsConnectionClosed(err) {
						close(connection.receiveLoopStopChannel)
						connection.Close()
						return
					}
					if connection.warningLogger != nil {
						connection.warningLogger.Log(Event.New("failed to receive message", err).Error())
					}
					continue
				}
				if err := connection.addMessageToChannel(messageBytes); err != nil {
					if connection.warningLogger != nil {
						connection.warningLogger.Log(Event.New("failed to add message to processing channel", err).Error())
					}
					connection.messageChannelSemaphore.ReleaseBlocking()
				}
			}

		}
	}
}

func (connection *TcpSystemgeConnection) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := connection.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncToken()) > maxSyncTokenSize {
		return Event.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Event.New("Message missing topic", nil)
	}
	if maxTopicSize := connection.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Event.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := connection.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Event.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
