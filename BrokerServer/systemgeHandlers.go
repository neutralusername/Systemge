package BrokerServer

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Event"
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
			return "", errors.New("unknown async topic \"" + topic + "\"")
		}
	}
	for _, topic := range topics {
		server.asyncTopicSubscriptions[topic][connection] = true
		server.connectionAsyncSubscriptions[connection][topic] = true
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
			return "", errors.New("unknown sync topic \"" + topic + "\"")
		}
	}
	for _, topic := range topics {
		server.syncTopicSubscriptions[topic][connection] = true
		server.connectionsSyncSubscriptions[connection][topic] = true
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
			return "", errors.New("unknown async topic \"" + topic + "\"")
		}
	}
	for _, topic := range topics {
		delete(server.asyncTopicSubscriptions[topic], connection)
		delete(server.connectionAsyncSubscriptions[connection], topic)
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
			return "", errors.New("unknown sync topic \"" + topic + "\"")
		}
	}
	for _, topic := range topics {
		delete(server.syncTopicSubscriptions[topic], connection)
		delete(server.connectionsSyncSubscriptions[connection], topic)
	}
	return "", nil
}

func (server *Server) handleAsyncPropagate(sendingConnection SystemgeConnection.SystemgeConnection, message *Message.Message) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for connection := range server.asyncTopicSubscriptions[message.GetTopic()] {
		if connection != sendingConnection {
			err := connection.AsyncMessage(message.GetTopic(), message.GetPayload())
			if err != nil {
				if server.warningLogger != nil {
					server.warningLogger.Log(Event.New("failed to send async message to client \""+connection.GetName(), nil).Error())
				}
			} else {
				server.asyncMessagesPropagated.Add(1)
			}
		}
	}
}

func (server *Server) handleSyncPropagate(connection SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	server.mutex.Lock()
	responseChannels := []<-chan *Message.Message{}
	waitgroup := sync.WaitGroup{}
	for client := range server.syncTopicSubscriptions[message.GetTopic()] {
		if client != connection {
			server.syncRequestsPropagated.Add(1)
			waitgroup.Add(1)
			go func(client SystemgeConnection.SystemgeConnection) {
				defer waitgroup.Done()
				responseChannel, err := client.SyncRequest(message.GetTopic(), message.GetPayload())
				if err != nil {
					if server.warningLogger != nil {
						server.warningLogger.Log(Event.New("failed to send sync request to client \""+client.GetName(), nil).Error())
					}
				} else {
					responseChannels = append(responseChannels, responseChannel)
					server.syncRequestsPropagated.Add(1)
				}
			}(client)
		}
	}
	server.mutex.Unlock()
	waitgroup.Wait()

	responses := []*Message.Message{}
	for _, responseChannel := range responseChannels {
		response := <-responseChannel
		if response != nil {
			responses = append(responses, response)
		}
	}
	if len(responses) == 0 {
		return "", errors.New("no responses")
	}
	return Message.SerializeMessages(responses), nil
}

func (server *Server) onSystemgeConnection(connection SystemgeConnection.SystemgeConnection) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	err := connection.StartMessageHandlingLoop(server.messageHandler, true)
	if err != nil {
		return err
	}
	server.connectionAsyncSubscriptions[connection] = make(map[string]bool)
	server.connectionsSyncSubscriptions[connection] = make(map[string]bool)
	return nil
}

func (server *Server) onSystemgeDisconnection(connection SystemgeConnection.SystemgeConnection) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for topic := range server.connectionAsyncSubscriptions[connection] {
		delete(server.asyncTopicSubscriptions[topic], connection)
	}
	delete(server.connectionAsyncSubscriptions, connection)
	for topic := range server.connectionsSyncSubscriptions[connection] {
		delete(server.syncTopicSubscriptions[topic], connection)
	}
	delete(server.connectionsSyncSubscriptions, connection)

}
