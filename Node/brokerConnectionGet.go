package Node

import "Systemge/Error"

func (node *Node) getBrokerConnectionForTopic(topic string) (*brokerConnection, error) {
	brokerConnection := node.getTopicBrokerConnection(topic)
	if brokerConnection == nil {
		endpoint, err := node.resolveBrokerForTopic(topic)
		if err != nil {
			return nil, Error.New("Error resolving broker address for topic \""+topic+"\"", err)
		}
		brokerConnection = node.getBrokerConnection(endpoint.GetAddress())
		if brokerConnection == nil {
			brokerConnection, err = node.connectToBroker(endpoint)
			if err != nil {
				return nil, Error.New("Error connecting to broker", err)
			}
			err = node.addBrokerConnection(brokerConnection)
			if err != nil {
				return nil, Error.New("Error adding broker connection", err)
			}
		}
		err = node.addTopicBrokerConnection(topic, brokerConnection)
		if err != nil {
			return nil, Error.New("Error adding topic endpoint", err)
		}
	}
	return brokerConnection, nil
}
