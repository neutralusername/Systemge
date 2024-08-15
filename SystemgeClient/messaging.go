package SystemgeClient

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *SystemgeClient) AsyncMessage(topic, payload string) error {
	return client.sendMessage(Message.NewAsync(topic, payload))
}

func (client *SystemgeClient) SyncMessage(topic, payload string) (*Message.Message, error) {
	synctoken, responseChannel := client.initResponseChannel()
	err := client.sendMessage(Message.NewSync(topic, payload, synctoken))
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

func (client *SystemgeClient) initResponseChannel() (string, chan *Message.Message) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	syncToken := client.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	for _, ok := client.syncResponseChannels[syncToken]; ok; {
		syncToken = client.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	}
	client.syncResponseChannels[syncToken] = make(chan *Message.Message, 1)
	return syncToken, client.syncResponseChannels[syncToken]
}

func (client *SystemgeClient) removeResponseChannel(syncToken string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if responseChannel, ok := client.syncResponseChannels[syncToken]; ok {
		close(responseChannel)
		delete(client.syncResponseChannels, syncToken)
	}
}

func (client *SystemgeClient) addSyncResponse(message *Message.Message) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if responseChannel, ok := client.syncResponseChannels[message.GetSyncTokenToken()]; ok {
		responseChannel <- message
		close(responseChannel)
		delete(client.syncResponseChannels, message.GetSyncTokenToken())
		return nil
	}
	return Error.New("No response channel found", nil)
}
