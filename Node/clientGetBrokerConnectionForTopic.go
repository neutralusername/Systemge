package Node

import "Systemge/Error"

func (client *Node) getBrokerConnectionForTopic(topic string) (*brokerConnection, error) {
	brokerConnection := client.GetTopicResolution(topic)
	if brokerConnection == nil {
		resolution, err := client.resolveBrokerForTopic(topic)
		if err != nil {
			return nil, Error.New("Error resolving broker address for topic \""+topic+"\"", err)
		}
		brokerConnection = client.getBrokerConnection(resolution.GetAddress())
		if brokerConnection == nil {
			brokerConnection, err = client.connectToBroker(resolution)
			if err != nil {
				return nil, Error.New("Error connecting to message broker server", err)
			}
			err = client.addBrokerConnection(brokerConnection)
			if err != nil {
				return nil, Error.New("Error adding server connection", err)
			}
		}
		err = client.addTopicResolution(topic, brokerConnection)
		if err != nil {
			return nil, Error.New("Error adding topic resolution", err)
		}
	}
	return brokerConnection, nil
}
