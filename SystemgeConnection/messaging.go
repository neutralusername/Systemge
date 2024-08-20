package SystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (connection *SystemgeConnection) AsyncMessage(topic, payload string) error {
	err := connection.send(Message.NewAsync(topic, payload).Serialize())
	if err != nil {
		return err
	}
	connection.asyncMessagesSent.Add(1)
	return nil
}

func (connection *SystemgeConnection) SyncRequest(topic, payload string) (*Message.Message, error) {
	synctoken, responseChannel := connection.InitResponseChannel()
	err := connection.send(Message.NewSync(topic, payload, synctoken).Serialize())
	if err != nil {
		connection.RemoveResponseChannel(synctoken)
		return nil, err
	}
	connection.syncRequestsSent.Add(1)
	if connection.config.SyncRequestTimeoutMs > 0 {
		timeout := time.NewTimer(time.Duration(connection.config.SyncRequestTimeoutMs) * time.Millisecond)
		select {
		case responseMessage := <-responseChannel:
			if responseMessage.GetTopic() == Message.TOPIC_SUCCESS {
				connection.syncSuccessResponsesReceived.Add(1)
			} else if responseMessage.GetTopic() == Message.TOPIC_FAILURE {
				connection.syncFailureResponsesReceived.Add(1)
			}
			return responseMessage, nil
		case <-connection.closeChannel:
			connection.noSyncResponseReceived.Add(1)
			connection.RemoveResponseChannel(synctoken)
			return nil, Error.New("SystemgeClient stopped before receiving response", nil)
		case <-timeout.C:
			connection.noSyncResponseReceived.Add(1)
			connection.RemoveResponseChannel(synctoken)
			return nil, Error.New("Timeout before receiving response", nil)
		}
	} else {
		select {
		case responseMessage := <-responseChannel:
			if responseMessage.GetTopic() == Message.TOPIC_SUCCESS {
				connection.syncSuccessResponsesReceived.Add(1)
			} else if responseMessage.GetTopic() == Message.TOPIC_FAILURE {
				connection.syncFailureResponsesReceived.Add(1)
			}
			return responseMessage, nil
		case <-connection.closeChannel:
			connection.noSyncResponseReceived.Add(1)
			connection.RemoveResponseChannel(synctoken)
			return nil, Error.New("SystemgeClient stopped before receiving response", nil)
		}
	}
}

func (connection *SystemgeConnection) AddSyncResponse(message *Message.Message) error {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	if responseChannel, ok := connection.syncResponseChannels[message.GetSyncTokenToken()]; ok {
		responseChannel <- message
		close(responseChannel)
		delete(connection.syncResponseChannels, message.GetSyncTokenToken())
		return nil
	}
	return Error.New("No response channel found", nil)
}

func (connection *SystemgeConnection) InitResponseChannel() (string, chan *Message.Message) {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	syncToken := connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	for _, ok := connection.syncResponseChannels[syncToken]; ok; {
		syncToken = connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	}
	connection.syncResponseChannels[syncToken] = make(chan *Message.Message, 1)
	return syncToken, connection.syncResponseChannels[syncToken]
}

func (connection *SystemgeConnection) RemoveResponseChannel(syncToken string) {
	connection.syncMutex.Lock()
	defer connection.syncMutex.Unlock()
	if responseChannel, ok := connection.syncResponseChannels[syncToken]; ok {
		close(responseChannel)
		delete(connection.syncResponseChannels, syncToken)
	}
}
