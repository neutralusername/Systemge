package SystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *SystemgeConnection) AsyncMessage(topic, payload string) error {
	return client.SendMessage(Message.NewAsync(topic, payload).Serialize())
}

func (client *SystemgeConnection) SyncMessage(topic, payload string) (*Message.Message, error) {
	synctoken, responseChannel := client.initResponseChannel()
	err := client.SendMessage(Message.NewSync(topic, payload, synctoken).Serialize())
	if err != nil {
		client.removeResponseChannel(synctoken)
		return nil, err
	}
	if client.config.SyncRequestTimeoutMs > 0 {
		timeout := time.NewTimer(time.Duration(client.config.SyncRequestTimeoutMs) * time.Millisecond)
		select {
		case responseMessage := <-responseChannel:
			return responseMessage, nil
		case <-client.stopChannel:
			client.removeResponseChannel(synctoken)
			return nil, Error.New("SystemgeClient stopped before receiving response", nil)
		case <-timeout.C:
			client.removeResponseChannel(synctoken)
			return nil, Error.New("Timeout before receiving response", nil)
		}
	} else {
		select {
		case responseMessage := <-responseChannel:
			return responseMessage, nil
		case <-client.stopChannel:
			client.removeResponseChannel(synctoken)
			return nil, Error.New("SystemgeClient stopped before receiving response", nil)
		}
	}
}

func (client *SystemgeConnection) initResponseChannel() (string, chan *Message.Message) {
	client.syncMutex.Lock()
	defer client.syncMutex.Unlock()
	syncToken := client.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	for _, ok := client.syncResponseChannels[syncToken]; ok; {
		syncToken = client.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	}
	client.syncResponseChannels[syncToken] = make(chan *Message.Message, 1)
	return syncToken, client.syncResponseChannels[syncToken]
}

func (client *SystemgeConnection) removeResponseChannel(syncToken string) {
	client.syncMutex.Lock()
	defer client.syncMutex.Unlock()
	if responseChannel, ok := client.syncResponseChannels[syncToken]; ok {
		close(responseChannel)
		delete(client.syncResponseChannels, syncToken)
	}
}

func (client *SystemgeConnection) addSyncResponse(message *Message.Message) error {
	client.syncMutex.Lock()
	defer client.syncMutex.Unlock()
	if responseChannel, ok := client.syncResponseChannels[message.GetSyncTokenToken()]; ok {
		responseChannel <- message
		close(responseChannel)
		delete(client.syncResponseChannels, message.GetSyncTokenToken())
		return nil
	}
	return Error.New("No response channel found", nil)
}
