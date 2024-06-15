package Client

import "Systemge/Error"

func (client *Client) getServerConnectionForTopic(topic string) (*serverConnection, error) {
	serverConnection := client.getTopicResolution(topic)
	if serverConnection == nil {
		broker, err := client.resolveBrokerForTopic(topic)
		if err != nil {
			return nil, Error.New("Error resolving broker address for topic \""+topic+"\"", err)
		}
		serverConnection = client.getServerConnection(broker.Address)
		if serverConnection == nil {
			serverConnection, err = client.connectToBroker(broker)
			if err != nil {
				return nil, Error.New("Error connecting to message broker server", err)
			}
			err = client.addServerConnection(serverConnection)
			if err != nil {
				return nil, Error.New("Error adding server connection", err)
			}
			err = client.addTopicResolution(topic, serverConnection)
			if err != nil {
				return nil, Error.New("Error adding topic resolution", err)
			}
		}
	}
	return serverConnection, nil
}

func (client *Client) attemptToReconnect(serverConnection *serverConnection) {
	client.mapOperationMutex.Lock()
	serverConnection.mapOperationMutex.Lock()
	delete(client.activeServerConnections, serverConnection.resolution.Address)
	topicsToReconnect := make([]string, 0)
	for topic := range serverConnection.topics {
		delete(client.topicResolutions, topic)
		if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
			topicsToReconnect = append(topicsToReconnect, topic)
		}
	}
	serverConnection.topics = make(map[string]bool)
	serverConnection.mapOperationMutex.Unlock()
	client.mapOperationMutex.Unlock()

	for _, topic := range topicsToReconnect {
		newServerConnection, err := client.getServerConnectionForTopic(topic)
		if err != nil {
			panic(Error.New("Unable to obtain new broker for topic \""+topic+"\"", err))
		}
		err = client.subscribeTopic(newServerConnection, topic)
		if err != nil {
			panic(Error.New("Unable to subscribe to topic \""+topic+"\"", err))
		}
	}
}
