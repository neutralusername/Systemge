package BrokerServer

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *MessageBrokerServer) subscribeAsync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown async topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		server.asyncTopicSubscriptions[topic][connection] = true
		server.asyncConnectionSubscriptions[connection][topic] = true
	}
	return "", nil
}

func (server *MessageBrokerServer) subscribeSync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown sync topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		server.syncTopicSubscriptions[topic][connection] = true
		server.syncConnectionSubscriptions[connection][topic] = true
	}
	return "", nil
}

func (server *MessageBrokerServer) unsubscribeAsync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown async topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		delete(server.asyncTopicSubscriptions[topic], connection)
		delete(server.asyncConnectionSubscriptions[connection], topic)
	}
	return "", nil
}

func (server *MessageBrokerServer) unsubscribeSync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown sync topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		delete(server.syncTopicSubscriptions[topic], connection)
		delete(server.syncConnectionSubscriptions[connection], topic)
	}
	return "", nil
}

func (server *MessageBrokerServer) handleAsyncPropagate(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for client := range server.asyncTopicSubscriptions[message.GetTopic()] {
		if client != connection {
			client.AsyncMessage(message.GetTopic(), message.GetPayload())
		}
	}
}

func (server *MessageBrokerServer) handleSyncPropagate(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	responseChannels := []<-chan *Message.Message{}
	for client := range server.syncTopicSubscriptions[message.GetTopic()] {
		if client != connection {
			responseChannel, err := client.SyncRequest(message.GetTopic(), message.GetPayload())
			if err != nil {
				if server.warningLogger != nil {
					server.warningLogger.Log(Error.New("failed to send sync request to client \""+client.GetName(), nil).Error())
				}
				continue
			}
			responseChannels = append(responseChannels, responseChannel)
		}
	}
	responses := []*Message.Message{}
	for _, responseChannel := range responseChannels {
		response := <-responseChannel
		if response != nil {
			responses = append(responses, response)
		}
	}
	if len(responses) == 0 {
		return "", Error.New("no responses", nil)
	}

	return Message.SerializeMessages(responses), nil
}

func (server *MessageBrokerServer) onSystemgeConnection(connection *SystemgeConnection.SystemgeConnection) error {
	server.mutex.Lock()
	server.asyncConnectionSubscriptions[connection] = make(map[string]bool)
	server.syncConnectionSubscriptions[connection] = make(map[string]bool)
	server.mutex.Unlock()
	connection.StartProcessingLoopSequentially(server.messageHandler)
	return nil
}

func (server *MessageBrokerServer) onSystemgeDisconnection(connection *SystemgeConnection.SystemgeConnection) {
	connection.StopProcessingLoop()
	server.mutex.Lock()
	for topic := range server.asyncConnectionSubscriptions[connection] {
		delete(server.asyncTopicSubscriptions[topic], connection)
	}
	delete(server.asyncConnectionSubscriptions, connection)
	for topic := range server.syncConnectionSubscriptions[connection] {
		delete(server.syncTopicSubscriptions[topic], connection)
	}
	delete(server.syncConnectionSubscriptions, connection)
	server.mutex.Unlock()
}
