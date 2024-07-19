package Node

import "Systemge/Error"

func (node *Node) getBrokerConnectionForTopic(topic string) *brokerConnection {
	brokerConnection := node.getTopicResolution(topic)
	if brokerConnection == nil {
		endpoint, err := node.resolveBrokerForTopic(topic)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed resolving broker address for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
			}
			return nil
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Resolved broker address \""+endpoint.Address+"\" for topic \""+topic+" \" on node \""+node.GetName()+"\"", nil).Error())
		}
		brokerConnection = node.getBrokerConnection(endpoint.Address)
		if brokerConnection == nil {
			brokerConnection, err = node.connectToBroker(endpoint)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed connecting to broker \""+endpoint.Address+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
				}
				return nil
			}
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Connected to broker \""+endpoint.Address+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
			}
			err = node.addBrokerConnection(brokerConnection)
			if err != nil {
				brokerConnection.close()
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed adding broker connection \""+endpoint.Address+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", err).Error())
				}
				return nil
			}
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Added broker connection \""+endpoint.Address+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
			}
		} else {
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Found existing broker connection \""+endpoint.Address+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
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
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed adding topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", err).Error())
			}
			return nil
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Added topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", nil).Error())
		}
	} else {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Found existing topic resolution \""+brokerConnection.endpoint.Address+"\" for topic \""+topic+"\" on node \""+node.GetName()+"\"", nil).Error())
		}
	}
	return brokerConnection
}
