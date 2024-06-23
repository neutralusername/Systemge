package Client

import (
	"Systemge/Utilities"
)

func (client *Client) getBrokerConnectionForTopic(topic string) (*brokerConnection, error) {
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
		err = client.addTopicResolution(topic, brokerConnection)
		if err != nil {
			return nil, Utilities.NewError("Error adding topic resolution", err)
		}
	}
	return brokerConnection, nil
}

func (client *Client) attemptToReconnectToSubscribedTopic(topic string) error {
	newBrokerConnection, err := client.getBrokerConnectionForTopic(topic)
	if err != nil {
		return Utilities.NewError("Unable to obtain new broker for topic \""+topic+"\"", err)
	}
	err = client.subscribeTopic(newBrokerConnection, topic)
	if err != nil {
		return Utilities.NewError("Unable to subscribe to topic \""+topic+"\"", err)
	}
	return nil
}
