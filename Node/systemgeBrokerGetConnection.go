package Node

import (
	"Systemge/Error"
)

func (node *Node) getBrokerConnectionForTopic(topic string, addTopicResolution bool) (*brokerConnection, error) {
	systemge := node.systemge
	if systemge == nil {
		return nil, Error.New("systemge component not initialized", nil)
	}
	brokerConnection := systemge.getTopicResolution(topic)
	if brokerConnection == nil {
		endpoint, err := systemge.resolveBrokerForTopic(node.GetName(), topic)
		if err != nil {
			return nil, Error.New("failed resolving broker address for topic \""+topic+"\"", err)
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Resolved broker address \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
		}
		brokerConnection = systemge.getBrokerConnection(endpoint.Address)
		if brokerConnection == nil {
			brokerConnection, err = systemge.connectToBroker(node.GetName(), endpoint)
			if err != nil {
				return nil, Error.New("failed connecting to broker \""+endpoint.Address+"\" for topic \""+topic+"\"", err)
			}
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Connected to broker \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
			}
			err = systemge.addBrokerConnection(brokerConnection)
			if err != nil {
				brokerConnection.close()
				return nil, Error.New("failed adding broker connection \""+endpoint.Address+"\" for topic \""+topic+"\"", err)
			}
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Added broker connection \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
			}
			go node.handleBrokerConnectionMessages(brokerConnection)
		} else {
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Found existing broker connection \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
			}
		}
		if addTopicResolution {
			err = systemge.addTopicResolution(topic, brokerConnection)
			if err != nil {
				brokerConnection.mutex.Lock()
				subscribedTopicsCount := len(brokerConnection.subscribedTopics)
				topicResolutionsCount := len(brokerConnection.topicResolutions)
				if subscribedTopicsCount == 0 && topicResolutionsCount == 0 {
					brokerConnection.close()
				}
				brokerConnection.mutex.Unlock()
				return nil, Error.New("failed adding topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.Address+"\"", err)
			}
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Added topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.Address+"\"", nil).Error())
			}
			go func() {
				err := systemge.removeTopicResolutionTimeout(topic, brokerConnection)
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to remove topic resolution for topic \""+topic+"\" from broker connection \""+brokerConnection.endpoint.Address+"\"", err).Error())
					}
				}
				if infoLogger := node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Removed topic resolution for topic \""+topic+"\" from broker connection \""+brokerConnection.endpoint.Address+"\"", nil).Error())
				}
			}()
		}
	} else {
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Found existing topic resolution \""+brokerConnection.endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
		}
	}
	return brokerConnection, nil
}
