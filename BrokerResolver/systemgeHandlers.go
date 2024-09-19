package BrokerResolver

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (resolver *Resolver) resolveAsync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolution, ok := resolver.asyncTopicTcpClientConfigs[message.GetPayload()]; ok {
		return Helpers.JsonMarshal(resolution), nil
	} else {
		return "", Error.New("Unkown async topic", nil)
	}
}

func (resolver *Resolver) resolveSync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolution, ok := resolver.syncTopicTcpClientConfigs[message.GetPayload()]; ok {
		return Helpers.JsonMarshal(resolution), nil
	} else {
		return "", Error.New("Unkown sync topic", nil)
	}
}

func (resolver *Resolver) onConnect(connection SystemgeConnection.SystemgeConnection) error {
	resolver.ongoingResolutions.Add(1)
	defer resolver.ongoingResolutions.Add(-1)
	message, err := connection.GetNextMessage()
	if err != nil {
		resolver.failedResolutions.Add(1)
		return err
	}
	switch message.GetTopic() {
	case Message.TOPIC_RESOLVE_ASYNC:
		err := connection.HandleMessage(message, resolver.messageHandler)
		if err != nil {
			resolver.failedResolutions.Add(1)
			return err
		}
		connection.Close()
		resolver.sucessfulAsyncResolutions.Add(1)
		return nil
	case Message.TOPIC_RESOLVE_SYNC:
		err := connection.HandleMessage(message, resolver.messageHandler)
		if err != nil {
			resolver.failedResolutions.Add(1)
			return err
		}
		connection.Close()
		resolver.sucessfulSyncResolutions.Add(1)
		return nil
	default:
		resolver.failedResolutions.Add(1)
		return Error.New("Invalid topic", nil)
	}
}
