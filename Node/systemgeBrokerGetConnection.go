package Node

import "Systemge/Error"

func (node *Node) getBrokerConnectionForTopic(topic string) *brokerConnection {
	brokerConnection := node.getTopicResolution(topic)
	if brokerConnection == nil {
		endpoint, err := node.resolveBrokerForTopic(topic)
		if err != nil {
			node.config.Logger.Warning(Error.New("Failed resolving broker address for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
			return nil
		} else {
			node.config.Logger.Info(Error.New("Resolved broker address \""+endpoint.GetAddress()+"\" for topic \""+topic+" \" on node \""+node.GetName()+"\"", nil).Error())
		}
		brokerConnection = node.getBrokerConnection(endpoint.GetAddress())
		if brokerConnection == nil {
			brokerConnection, err = node.connectToBroker(endpoint)
			if err != nil {
				node.config.Logger.Warning(Error.New("Failed connecting to broker \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
				return nil
			} else {
				node.config.Logger.Info(Error.New("Connected to broker \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
			}
			err = node.addBrokerConnection(brokerConnection)
			if err != nil {
				brokerConnection.close()
				node.config.Logger.Warning(Error.New("Failed adding broker connection \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
				return nil
			} else {
				node.config.Logger.Info(Error.New("Added broker connection \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
			}
		} else {
			node.config.Logger.Info(Error.New("Found existing broker connection \""+endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
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
			node.config.Logger.Warning(Error.New("Failed adding topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.GetName()+"\"", err).Error())
			return nil
		} else {
			node.config.Logger.Info(Error.New("Added topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.GetName()+"\"", nil).Error())
		}
	} else {
		node.config.Logger.Info(Error.New("Found existing topic resolution \""+brokerConnection.endpoint.GetAddress()+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
	}
	return brokerConnection
}
