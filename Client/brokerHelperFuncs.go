package Client

import (
	"Systemge/Utilities"
)

func (client *Client) getBrokerConnectionForTopic(topic string) (*brokerConnection, error) {
	brokerConnection := client.getTopicResolution(topic)
	if brokerConnection == nil {
		broker, err := client.resolveBrokerForTopic(topic)
		if err != nil {
			return nil, Utilities.NewError("Error resolving broker address for topic \""+topic+"\"", err)
		}
		brokerConnection = client.getBrokerConnection(broker.Address)
		if brokerConnection == nil {
			brokerConnection, err = client.connectToBroker(broker)
			if err != nil {
				return nil, Utilities.NewError("Error connecting to message broker server", err)
			}
			err = client.addBrokerConnection(brokerConnection)
			if err != nil {
				return nil, Utilities.NewError("Error adding server connection", err)
			}
			err = client.addTopicResolution(topic, brokerConnection)
			if err != nil {
				return nil, Utilities.NewError("Error adding topic resolution", err)
			}
		}
	}
	return brokerConnection, nil
}

func (client *Client) attemptToReconnect(brokerConnection *brokerConnection) {
	client.mapOperationMutex.Lock()
	brokerConnection.mapOperationMutex.Lock()
	delete(client.activeBrokerConnections, brokerConnection.resolution.Address)
	topicsToReconnect := make([]string, 0)
	for topic := range brokerConnection.topics {
		delete(client.topicResolutions, topic)
		if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
			topicsToReconnect = append(topicsToReconnect, topic)
		}
	}
	brokerConnection.topics = make(map[string]bool)
	brokerConnection.mapOperationMutex.Unlock()
	client.mapOperationMutex.Unlock()
	if client.isStarted {
		for _, topic := range topicsToReconnect {
			newBrokerConnection, err := client.getBrokerConnectionForTopic(topic)
			if err != nil {
				panic(Utilities.NewError("Unable to obtain new broker for topic \""+topic+"\"", err))
			}
			err = client.subscribeTopic(newBrokerConnection, topic)
			if err != nil {
				panic(Utilities.NewError("Unable to subscribe to topic \""+topic+"\"", err))
			}
		}
	}
}
