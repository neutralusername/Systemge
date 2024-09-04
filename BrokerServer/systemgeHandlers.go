package BrokerServer

import (
	"encoding/json"
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *Server) subscribeAsync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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

func (server *Server) subscribeSync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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

func (server *Server) unsubscribeAsync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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

func (server *Server) unsubscribeSync(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
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

func (server *Server) handleAsyncPropagate(connection SystemgeConnection.SystemgeConnection, message *Message.Message) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for client := range server.asyncTopicSubscriptions[message.GetTopic()] {
		if client != connection {
			client.AsyncMessage(message.GetTopic(), message.GetPayload())
		}
	}
}

func (server *Server) handleSyncPropagate(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {

	server.mutex.Lock()
	responseChannels := []<-chan *Message.Message{}
	waitgroup := sync.WaitGroup{}
	for client := range server.syncTopicSubscriptions[message.GetTopic()] {
		if client != connection {
			waitgroup.Add(1)
			go func(client SystemgeConnection.SystemgeConnection) {
				defer waitgroup.Done()
				responseChannel, err := client.SyncRequest(message.GetTopic(), message.GetPayload())
				if err != nil {
					if server.warningLogger != nil {
						server.warningLogger.Log(Error.New("failed to send sync request to client \""+client.GetName(), nil).Error())
					}
					responseChannels = append(responseChannels, nil)
					return
				}
				responseChannels = append(responseChannels, responseChannel)
			}(client)
		}
	}
	server.mutex.Unlock()
	waitgroup.Wait()

	responses := []*Message.Message{}
	for _, responseChannel := range responseChannels {
		if responseChannel == nil {
			continue
		}
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

func (server *Server) onSystemgeConnection(connection SystemgeConnection.SystemgeConnection) error {
	server.mutex.Lock()
	server.asyncConnectionSubscriptions[connection] = make(map[string]bool)
	server.syncConnectionSubscriptions[connection] = make(map[string]bool)
	server.mutex.Unlock()
	connection.StartProcessingLoopSequentially(server.messageHandler)
	return nil
}

func (server *Server) onSystemgeDisconnection(connection SystemgeConnection.SystemgeConnection) {
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
