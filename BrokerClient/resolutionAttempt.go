package BrokerClient

import (
	"time"

	"github.com/neutralusername/systemge/Config"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/tools"
)

type resolutionAttempt struct {
	topic       string
	isSyncTopic bool
	ongoing     chan bool
	connections map[string]*connection
}

func (messageBrokerClient *Client) startResolutionAttempt(topic string, syncTopic bool, stopChannel chan bool) (*resolutionAttempt, error) {
	if stopChannel != messageBrokerClient.stopChannel {
		return nil, Event.New("Aborted because resolution attempt belongs to outdated session", nil)
	}
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()

	if resolutionAttempt := messageBrokerClient.ongoingTopicResolutions[topic]; resolutionAttempt != nil {
		return resolutionAttempt, nil
	}
	resolutionAttempt := &resolutionAttempt{
		ongoing:     make(chan bool),
		topic:       topic,
		isSyncTopic: syncTopic,
	}
	messageBrokerClient.ongoingTopicResolutions[topic] = resolutionAttempt
	messageBrokerClient.waitGroup.Add(1)
	messageBrokerClient.resolutionAttempts.Add(1)
	go func() {
		messageBrokerClient.resolutionAttempt(resolutionAttempt, stopChannel)
	}()
	return resolutionAttempt, nil
}

func (messageBrokerClient *Client) resolutionAttempt(resolutionAttempt *resolutionAttempt, stopChannel chan bool) {
	var tcpClientConfigs []*Config.TcpClient
	attempts := uint32(0)
	for len(tcpClientConfigs) == 0 && stopChannel == messageBrokerClient.stopChannel && (messageBrokerClient.config.ResolutionMaxAttempts == 0 || attempts < messageBrokerClient.config.ResolutionMaxAttempts) {
		tcpClientConfigs = messageBrokerClient.resolveBrokerTcpClientConfigs(resolutionAttempt.topic, resolutionAttempt.isSyncTopic)

		messageBrokerClient.mutex.Lock()
		subscribe := (resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedSyncTopics[resolutionAttempt.topic]) || (!resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedAsyncTopics[resolutionAttempt.topic])
		messageBrokerClient.mutex.Unlock()
		if !subscribe {
			break
		}
		attempts++
		select {
		case <-time.After(time.Duration(messageBrokerClient.config.ResolutionAttemptRetryIntervalMs) * time.Millisecond):
		case <-stopChannel:
		}
	}

	connections := map[string]*connection{}
	for _, tcpClientConfig := range tcpClientConfigs {
		conn, err := messageBrokerClient.getBrokerConnection(tcpClientConfig, stopChannel)
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Event.New("Failed to get connection to resolved tcpClientConfig \""+tcpClientConfig.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(tools.NewMail(nil, "error", Event.New("Failed to get connection to resolved tcpClientConfig \""+tcpClientConfig.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Event.New("Failed to send email", err).Error())
					}
				}
			}
			continue
		}
		messageBrokerClient.mutex.Lock()
		if resolutionAttempt.isSyncTopic {
			conn.responsibleSyncTopics[resolutionAttempt.topic] = true
		} else {
			conn.responsibleAsyncTopics[resolutionAttempt.topic] = true
		}
		connections[getTcpClientConfigString(tcpClientConfig)] = conn
		subscribe := (resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedSyncTopics[resolutionAttempt.topic]) || (!resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedAsyncTopics[resolutionAttempt.topic])
		messageBrokerClient.mutex.Unlock()

		if subscribe {
			messageBrokerClient.subscribeToTopic(conn, resolutionAttempt.topic, resolutionAttempt.isSyncTopic)
		}
	}

	messageBrokerClient.mutex.Lock()
	for tcpClientConfigString, existingConnection := range messageBrokerClient.topicResolutions[resolutionAttempt.topic] {
		if _, ok := connections[tcpClientConfigString]; !ok {
			if resolutionAttempt.isSyncTopic {
				delete(existingConnection.responsibleSyncTopics, resolutionAttempt.topic)
			} else {
				delete(existingConnection.responsibleAsyncTopics, resolutionAttempt.topic)
			}
			if len(existingConnection.responsibleAsyncTopics) == 0 && len(existingConnection.responsibleSyncTopics) == 0 {
				delete(messageBrokerClient.brokerConnections, tcpClientConfigString)
				existingConnection.connection.Close()
			}
		}
	}
	if len(connections) > 0 {
		messageBrokerClient.topicResolutions[resolutionAttempt.topic] = connections
		go messageBrokerClient.handleTopicResolutionLifetime(resolutionAttempt.topic, resolutionAttempt.isSyncTopic, stopChannel)
	}
	resolutionAttempt.connections = connections
	delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
	close(resolutionAttempt.ongoing)
	messageBrokerClient.mutex.Unlock()

	messageBrokerClient.waitGroup.Done()
}
