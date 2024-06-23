package Client

import (
	"Systemge/Utilities"
)

func (client *Client) getBrokerConnectionForTopic(topic string, addTopicResolution bool) (*brokerConnection, error) {
	brokerConnection := client.getTopicResolution(topic)
	if brokerConnection == nil {
		resolution, err := client.resolveBrokerForTopic(client.resolverResolution, topic)
		if err != nil {
			return nil, Utilities.NewError("Error resolving broker address for topic \""+topic+"\"", err)
		}
		brokerConnection = client.getBrokerConnection(resolution.GetAddress())
		if brokerConnection == nil {
			brokerConnection, err = client.connectToBroker(resolution)
			if err != nil {
				return nil, Utilities.NewError("Error connecting to message broker server", err)
			}
			err = client.addBrokerConnection(brokerConnection)
			if err != nil {
				return nil, Utilities.NewError("Error adding server connection", err)
			}
		}
		if addTopicResolution {
			err = client.addTopicResolution(topic, brokerConnection)
			if err != nil {
				return nil, Utilities.NewError("Error adding topic resolution", err)
			}
		}
	}
	return brokerConnection, nil
}

func (client *Client) attemptToReconnect(brokerConnection *brokerConnection) error {
	client.mapOperationMutex.Lock()
	brokerConnection.mutex.Lock()
	delete(client.activeBrokerConnections, brokerConnection.resolution.GetAddress())
	topicsToReconnect := make([]string, 0)
	for topic := range brokerConnection.topics {
		delete(client.topicResolutions, topic)
		if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
			topicsToReconnect = append(topicsToReconnect, topic)
		}
	}
	brokerConnection.topics = make(map[string]bool)
	brokerConnection.mutex.Unlock()
	client.mapOperationMutex.Unlock()
	for _, topic := range topicsToReconnect {
		newBrokerConnection, err := client.getBrokerConnectionForTopic(topic, false)
		if err != nil {
			return Utilities.NewError("Unable to obtain new broker for topic \""+topic+"\"", err)
		}
		err = client.subscribeTopic(newBrokerConnection, topic)
		if err != nil {
			return Utilities.NewError("Unable to subscribe to topic \""+topic+"\"", err)
		}
	}
	return nil
}

func (client *Client) removeAllBrokerConnections() {
	client.mapOperationMutex.Lock()
	for address, brokerConnection := range client.activeBrokerConnections {
		brokerConnection.close()
		delete(client.activeBrokerConnections, address)
		for topic := range brokerConnection.topics {
			delete(client.topicResolutions, topic)
		}
	}
	client.mapOperationMutex.Unlock()
}
