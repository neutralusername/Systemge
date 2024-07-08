package Node

import "Systemge/Error"

func (node *Node) getBrokerConnectionForTopic(topic string) (*brokerConnection, error) {
	brokerConnection := node.getTopicBrokerConnection(topic)
	if brokerConnection == nil {
		endpoint, err := node.resolveBrokerForTopic(topic)
		if err != nil {
			return nil, Error.New("failed resolving broker address for topic \""+topic+"\"", err)
		}
		brokerConnection = node.getBrokerConnection(endpoint.GetAddress())
		if brokerConnection == nil {
			brokerConnection, err = node.connectToBroker(endpoint)
			if err != nil {
				return nil, Error.New("failed connecting to broker \""+endpoint.GetAddress()+"\"", err)
			}
			err = node.addBrokerConnection(brokerConnection)
			if err != nil {
				brokerConnection.close()
				return nil, Error.New("failed adding broker connection \""+endpoint.GetAddress()+"\"", err)
			}
		}
		err = node.addTopicResolution(topic, brokerConnection)
		if err != nil {
			brokerConnection.mutex.Lock()
			subscribedTopicsCount := len(brokerConnection.subscribedTopics)
			topicResolutionsCount := len(brokerConnection.topicResolutions)
			brokerConnection.mutex.Unlock()
			if subscribedTopicsCount == 0 && topicResolutionsCount == 0 {
				node.handleBrokerDisconnect(brokerConnection)
			}
			return nil, Error.New("failed adding topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.GetAddress()+"\"", err)
		}
	}
	return brokerConnection, nil
}
