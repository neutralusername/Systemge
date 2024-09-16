package BrokerClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

type resolutionAttempt struct {
	topic       string
	isSyncTopic bool
	ongoing     chan bool
	endpoints   []*Config.TcpClient
}

func (messageBrokerClient *Client) startResolutionAttempt(topic string, syncTopic bool) (*resolutionAttempt, error) {
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
		resolutionAttempt.endpoints = messageBrokerClient.resolveBrokerEndpoints(resolutionAttempt.topic, resolutionAttempt.isSyncTopic)
		messageBrokerClient.mutex.Lock()
		delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
		messageBrokerClient.mutex.Unlock()
		close(resolutionAttempt.ongoing)
	}()
	return resolutionAttempt, nil
}

func (server *Client) addSubscribedTopicConnectionAttempts(endpoints []*Config.TcpClient, topic string) {
	for _, endpoint := range endpoints {
		err := server.subscribedTopicSystemgeClient.AddConnectionAttempt(endpoint)
		if err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Error.New("Failed to add connectionAttempt to endpoint \""+endpoint.Address+"\" for topic \""+topic+"\"", err).Error())
			}
			if server.mailer != nil {
				if err := server.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to add connectionAttempt to endpoint \""+endpoint.Address+"\" for topic \""+topic+"\"", err).Error())); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
		}
	}
}
