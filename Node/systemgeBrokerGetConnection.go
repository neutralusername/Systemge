package Node

import (
	"github.com/neutralusername/Systemge/Error"
)

func (systemge *systemgeComponent) resolutionSafetyMechanism(topic string) (*brokerConnection, bool) {
	systemge.topicResolutionMutex.Lock()
	if systemge.topicResolutions[topic] != nil {
		systemge.topicResolutionMutex.Unlock()
		return systemge.topicResolutions[topic], false
	} else {
		chanLock, ok := systemge.topicsCurrentlyBeingResolved[topic]
		if !ok {
			systemge.topicsCurrentlyBeingResolved[topic] = make(chan struct{})
			systemge.topicResolutionMutex.Unlock()
			return nil, true
		} else {
			systemge.topicResolutionMutex.Unlock()
			<-chanLock
			systemge.topicResolutionMutex.Lock()
			brokerConnection := systemge.topicResolutions[topic]
			systemge.topicResolutionMutex.Unlock()
			return brokerConnection, false
		}
	}
}

func (node *Node) getBrokerConnectionForTopic(topic string, addTopicResolution bool) (*brokerConnection, error) {
	systemge := node.systemge
	if systemge == nil {
		return nil, Error.New("systemge component not initialized", nil)
	}
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Getting topic resolution for topic \""+topic+"\"", nil).Error())
	}
	brokerConnection, def := systemge.resolutionSafetyMechanism(topic)
	if brokerConnection != nil {
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Found existing topic resolution for topic \""+topic+"\"", nil).Error())
		}
		return brokerConnection, nil
	} else {
		if def {
			defer func() {
				systemge.topicResolutionMutex.Lock()
				close(systemge.topicsCurrentlyBeingResolved[topic])
				delete(systemge.topicsCurrentlyBeingResolved, topic)
				systemge.topicResolutionMutex.Unlock()
			}()
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("No existing topic resolution found for topic \""+topic+"\". Resolving broker address", nil).Error())
		}
		endpoint, err := systemge.resolveBrokerForTopic(node.GetName(), topic)
		if err != nil {
			return nil, Error.New("Failed resolving broker address for topic \""+topic+"\"", err)
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Resolved broker address for topic \""+topic+"\". Getting existing broker connection for \""+endpoint.Address+"\"", nil).Error())
		}
		brokerConnection = systemge.getBrokerConnection(endpoint.Address)
		if brokerConnection == nil {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("No existing broker connection found for \""+endpoint.Address+"\". Connecting to broker for topic \""+topic+"\"", nil).Error())
			}
			brokerConnection, err = systemge.connectToBroker(node.GetName(), endpoint)
			if err != nil {
				return nil, Error.New("Failed connecting to broker \""+endpoint.Address+"\" for topic \""+topic+"\"", err)
			}
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Connected to broker \""+endpoint.Address+"\" for topic \""+topic+"\". Adding broker connection", nil).Error())
			}
			err = systemge.addBrokerConnection(brokerConnection)
			if err != nil {
				brokerConnection.closeNetConn(true)
				return nil, Error.New("Failed adding broker connection \""+endpoint.Address+"\" for topic \""+topic+"\". Closed broker connection", err)
			}
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Added broker connection \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
			}
			go node.handleBrokerConnectionMessages(brokerConnection)
		} else {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Found existing broker connection \""+endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
			}
		}
		if addTopicResolution && systemge.application.GetSystemgeComponentConfig().TopicResolutionLifetimeMs != 0 {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Adding topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.Address+"\"", nil).Error())
			}
			err = systemge.addTopicResolution(topic, brokerConnection)
			if err != nil {
				if brokerConnection.closeIfNoTopics() {
					if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
						infoLogger.Log(Error.New("Closed broker connection \""+brokerConnection.endpoint.Address+"\" due to no topics", nil).Error())
					}
				}
				return nil, Error.New("Failed adding topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.Address+"\"", err)
			}
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Added topic resolution for topic \""+topic+"\" to broker connection \""+brokerConnection.endpoint.Address+"\"", nil).Error())
			}
			if systemge.application.GetSystemgeComponentConfig().TopicResolutionLifetimeMs > 0 {
				go func() {
					if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
						infoLogger.Log(Error.New("Starting topic resolution lifetime loop for topic \""+topic+"\" on broker connection \""+brokerConnection.endpoint.Address+"\"", nil).Error())
					}
					defer func() {
						if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
							infoLogger.Log(Error.New("Stopped topic resolution lifetime loop for topic \""+topic+"\" on broker connection \""+brokerConnection.endpoint.Address+"\"", nil).Error())
						}
					}()
					for systemge == node.systemge {
						err := systemge.topicResolutionLifetimeTimeout(node.GetName(), topic, brokerConnection)
						if err != nil {
							if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
								warningLogger.Log(err.Error())
							}
							systemge.topicResolutionMutex.Lock()
							defer systemge.topicResolutionMutex.Unlock()
							delete(systemge.topicResolutions, topic)
							brokerConnection.removeTopicResolution(topic)
							if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
								infoLogger.Log(Error.New("Removed topic resolution for topic \""+topic+"\" from broker connection \""+brokerConnection.endpoint.Address+"\"", nil).Error())
							}
							if brokerConnection.closeIfNoTopics() {
								if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
									infoLogger.Log(Error.New("Closed broker connection \""+brokerConnection.endpoint.Address+"\" due to no topics", nil).Error())
								}
							}
							return
						} else {
							if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
								infoLogger.Log(Error.New("Resolved same broker address for topic \""+topic+"\". Resetting topic resolution timer", nil).Error())
							}
						}
					}
				}()
			}
		}
	}
	return brokerConnection, nil
}
