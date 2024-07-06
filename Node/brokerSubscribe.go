package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
)

func (node *Node) subscribeLoop(topic string) {
	for node.IsStarted() {
		node.logger.Log("Attempting connection for topic \"" + topic + "\"")

		endpoint, err := node.resolveBrokerForTopic(topic)
		if err != nil {
			node.logger.Log(Error.New("Unable to resolve broker for topic \""+topic+"\"", err).Error())
			continue
		}
		brokerConnection := node.getBrokerConnection(endpoint.GetAddress())
		if brokerConnection == nil {
			brokerConnection, err = node.connectToBroker(endpoint)
			if err != nil {
				node.logger.Log(Error.New("Unable to connect to broker for topic \""+topic+"\"", err).Error())
				continue
			}
			err = node.addBrokerConnection(brokerConnection)
			if err != nil {
				brokerConnection.close()
				node.logger.Log(Error.New("Unable to add broker connection for topic \""+topic+"\"", err).Error())
				continue
			}
		}
		err = node.subscribeTopic(brokerConnection, topic)
		if err != nil {
			node.logger.Log(Error.New("Unable to subscribe to topic \""+topic+"\"", err).Error())
			continue
		}
		err = brokerConnection.addSubscribedTopic(topic)
		if err != nil {
			node.logger.Log(Error.New("Unable to add topic to broker connection", err).Error())
			continue
		}
		node.logger.Log("connection for topic \"" + topic + "\" successful")
		break
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
