package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"time"
)

func (node *Node) subscribeLoop(topic string) {
	for node.IsStarted() {
		success := func() bool {
			node.logger.Log("Attempting connection for topic \"" + topic + "\"")
			endpoint, err := node.resolveBrokerForTopic(topic)
			if err != nil {
				node.logger.Log(Error.New("Unable to resolve broker for topic \""+topic+"\"", err).Error())
				return false
			}
			brokerConnection := node.getBrokerConnection(endpoint.GetAddress())
			if brokerConnection == nil {
				brokerConnection, err = node.connectToBroker(endpoint)
				if err != nil {
					node.logger.Log(Error.New("Unable to connect to broker for topic \""+topic+"\"", err).Error())
					return false
				}
				err = node.addBrokerConnection(brokerConnection)
				if err != nil {
					brokerConnection.close()
					node.logger.Log(Error.New("Unable to add broker connection for topic \""+topic+"\"", err).Error())
					return false
				}
			}
			err = node.subscribeTopic(brokerConnection, topic)
			if err != nil {
				node.logger.Log(Error.New("Unable to subscribe to topic \""+topic+"\"", err).Error())
				return false
			}
			err = brokerConnection.addSubscribedTopic(topic)
			if err != nil {
				brokerConnection.mutex.Lock()
				subscribedTopicsCount := len(brokerConnection.subscribedTopics)
				brokerConnection.mutex.Unlock()
				if subscribedTopicsCount == 0 {
					node.handleBrokerDisconnect(brokerConnection)
				}
				node.logger.Log(Error.New("Unable to add topic to broker connection", err).Error())
				return false
			}
			node.logger.Log("connection for topic \"" + topic + "\" successful")
			return true
		}()
		if success {
			break
		}
		time.Sleep(time.Duration(node.config.BrokerReconnectDelayMs) * time.Millisecond)
	}
}

func (node *Node) subscribeTopic(brokerConnection *brokerConnection, topic string) error {
	message := Message.NewSync("subscribe", node.config.Name, topic, node.randomizer.GenerateRandomString(10, Utilities.ALPHA_NUMERIC))
	responseChannel, err := node.addMessageWaitingForResponse(message)
	if err != nil {
		return Error.New("Error adding message to waiting for response map", err)
	}
	err = brokerConnection.send(message)
	if err != nil {
		node.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to broker", err)
	}
	response, err := node.receiveSyncResponse(message, responseChannel)
	if err != nil {
		return Error.New("Error sending subscription sync request", err)
	}
	if response.GetTopic() != "subscribed" {
		return Error.New("Invalid response topic \""+response.GetTopic()+"\"", nil)
	}
	return nil
}
