package BrokerClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/TcpConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type resolutionAttempt struct {
	topic       string
	isSyncTopic bool
	ongoing     chan bool
	err         error
}

func (messageBrokerClient *MessageBrokerClient) startResolutionAttempt(topic string, syncTopic bool, stopChannel chan bool) error {
	if stopChannel != messageBrokerClient.stopChannel {
		return Error.New("Aborted because resolution attempt is from previous session", nil)
	}
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	messageBrokerClient.waitGroup.Add(1)
	if resolutionAttempt := messageBrokerClient.ongoingTopicResolutions[topic]; resolutionAttempt != nil {
		<-resolutionAttempt.ongoing
		return resolutionAttempt.err
	}
	resolutionAttempt := &resolutionAttempt{
		ongoing:     make(chan bool),
		topic:       topic,
		isSyncTopic: syncTopic,
	}
	messageBrokerClient.ongoingTopicResolutions[topic] = resolutionAttempt

	go func() {
		messageBrokerClient.resolutionAttempt(resolutionAttempt, stopChannel)
		if resolutionAttempt.err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Error.New("Failed to resolve connection", resolutionAttempt.err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to resolve connection", resolutionAttempt.err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
		}
	}()
	return nil
}

func (messageBrokerClient *MessageBrokerClient) resolutionAttempt(resolutionAttempt *resolutionAttempt, stopChannel chan bool) {

	endpoints, err := messageBrokerClient.resolveBrokerEndpoints(resolutionAttempt.topic)
	if err != nil {
		resolutionAttempt.err = Error.New("Failed to resolve broker endpoints", err)
		return
	}

	connections := map[string]*connection{}
	for _, endpoint := range endpoints {
		conn := messageBrokerClient.brokerConnections[getEndpointString(endpoint)]
		if conn == nil {
			systemgeConnection, err := TcpConnection.EstablishConnection(messageBrokerClient.config.ConnectionConfig, endpoint, messageBrokerClient.GetName(), messageBrokerClient.config.MaxServerNameLength)
			if err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to establish connection to resolved endpoint \""+endpoint.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())
				}
				if messageBrokerClient.mailer != nil {
					if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to establish connection to resolved endpoint \""+endpoint.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())); err != nil {
						if messageBrokerClient.errorLogger != nil {
							messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
						}
					}
				}
				continue
			}
			conn = &connection{
				connection:             systemgeConnection,
				endpoint:               endpoint,
				responsibleAsyncTopics: make(map[string]bool),
				responsibleSyncTopics:  make(map[string]bool),
			}
		}
		if resolutionAttempt.isSyncTopic {
			conn.responsibleSyncTopics[resolutionAttempt.topic] = true
		} else {
			conn.responsibleAsyncTopics[resolutionAttempt.topic] = true
		}
		connections[getEndpointString(endpoint)] = conn
		if (resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedSyncTopics[resolutionAttempt.topic]) || (!resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedAsyncTopics[resolutionAttempt.topic]) {
			messageBrokerClient.subscribeToTopic(conn, resolutionAttempt.topic, resolutionAttempt.isSyncTopic)
		}
		messageBrokerClient.mutex.Lock()
		if messageBrokerClient.brokerConnections[getEndpointString(endpoint)] == nil {
			messageBrokerClient.brokerConnections[getEndpointString(endpoint)] = conn
			go messageBrokerClient.handleConnectionLifetime(conn, stopChannel)
		}
		messageBrokerClient.mutex.Unlock()
	}

	messageBrokerClient.mutex.Lock()
	for endpointString, existingConnection := range messageBrokerClient.topicResolutions[resolutionAttempt.topic] {
		if _, ok := connections[endpointString]; !ok {
			if resolutionAttempt.isSyncTopic {
				delete(existingConnection.responsibleSyncTopics, resolutionAttempt.topic)
			} else {
				delete(existingConnection.responsibleAsyncTopics, resolutionAttempt.topic)
			}
			if len(existingConnection.responsibleAsyncTopics) == 0 && len(existingConnection.responsibleSyncTopics) == 0 {
				delete(messageBrokerClient.brokerConnections, endpointString)
				existingConnection.connection.Close()
			}
		}
	}
	messageBrokerClient.topicResolutions[resolutionAttempt.topic] = connections

	delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
	close(resolutionAttempt.ongoing)
	messageBrokerClient.mutex.Unlock()
	messageBrokerClient.waitGroup.Done()

	go messageBrokerClient.handleTopicResolutionLifetime(resolutionAttempt.topic, resolutionAttempt.isSyncTopic, stopChannel)
}
