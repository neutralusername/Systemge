package Node

import "Systemge/Error"

func (node *Node) getBrokerConnectionForTopic(topic string) (*brokerConnection, error) {
	brokerConnection := node.GetTopicResolution(topic)
	if brokerConnection == nil {
		resolution, err := node.resolveBrokerForTopic(topic)
		if err != nil {
			return nil, Error.New("Error resolving broker address for topic \""+topic+"\"", err)
		}
		brokerConnection = node.getBrokerConnection(resolution.GetAddress())
		if brokerConnection == nil {
			brokerConnection, err = node.connectToBroker(resolution)
			if err != nil {
				return nil, Error.New("Error connecting to broker", err)
			}
			err = node.addBrokerConnection(brokerConnection)
			if err != nil {
				return nil, Error.New("Error adding broker connection", err)
			}
		}
		err = node.addTopicResolution(topic, brokerConnection)
		if err != nil {
			return nil, Error.New("Error adding topic resolution", err)
		}
	}
	return brokerConnection, nil
}
