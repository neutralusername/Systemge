package BrokerResolver

import (
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/helpers"
)

func (resolver *Resolver) resolveAsync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolution, ok := resolver.asyncTopicTcpClientConfigs[message.GetPayload()]; ok {
		return helpers.JsonMarshal(resolution), nil
	} else {
		return "", Event.New("Unkown async topic", nil)
	}
}

func (resolver *Resolver) resolveSync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolution, ok := resolver.syncTopicTcpClientConfigs[message.GetPayload()]; ok {
		return helpers.JsonMarshal(resolution), nil
	} else {
		return "", Event.New("Unkown sync topic", nil)
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
		return Event.New("Invalid topic", nil)
	}
}
