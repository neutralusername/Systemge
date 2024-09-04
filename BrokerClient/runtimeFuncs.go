package BrokerClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

func (messageBrokerClient *Client) ResolveSubscribeTopics() error {
	messageBrokerClient.statusMutex.Lock()
	if messageBrokerClient.status != Status.STARTED {
		messageBrokerClient.statusMutex.Unlock()
		return Error.New("Client is not started", nil)
	}

	resolutionAttempts := []*resolutionAttempt{}
	for asyncTopic := range messageBrokerClient.subscribedAsyncTopics {
		resolutionAttempt, _ := messageBrokerClient.startResolutionAttempt(asyncTopic, false, messageBrokerClient.stopChannel, messageBrokerClient.subscribedAsyncTopics[asyncTopic])
		resolutionAttempts = append(resolutionAttempts, resolutionAttempt)
	}
	for syncTopic := range messageBrokerClient.subscribedSyncTopics {
		resolutionAttempt, _ := messageBrokerClient.startResolutionAttempt(syncTopic, true, messageBrokerClient.stopChannel, messageBrokerClient.subscribedSyncTopics[syncTopic])
		resolutionAttempts = append(resolutionAttempts, resolutionAttempt)
	}
	messageBrokerClient.statusMutex.Unlock()

	for _, resolutionAttempt := range resolutionAttempts {
		<-resolutionAttempt.ongoing
	}
	return nil
}

func (messageBrokerClient *Client) ResolveTopic(topic string) error {
	messageBrokerClient.statusMutex.Lock()
	if messageBrokerClient.status != Status.STARTED {
		messageBrokerClient.statusMutex.Unlock()
		return Error.New("Client is not started", nil)
	}
	resolutionAttempt, _ := messageBrokerClient.startResolutionAttempt(topic, true, messageBrokerClient.stopChannel, (messageBrokerClient.subscribedSyncTopics[topic] || messageBrokerClient.subscribedAsyncTopics[topic]))
	messageBrokerClient.statusMutex.Unlock()
	<-resolutionAttempt.ongoing
	return nil
}

func (messageBrokerClient *Client) AddAsyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	messageBrokerClient.subscribedAsyncTopics[topic] = true
	return nil
}

func (messageBrokerClient *Client) AddSyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	messageBrokerClient.subscribedSyncTopics[topic] = true
	return nil
}

func (messageBrokerClient *Client) RemoveAsyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	delete(messageBrokerClient.subscribedAsyncTopics, topic)
	return nil
}

func (messageBrokerClient *Client) RemoveSyncSubscribeTopic(topic string) error {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()
	delete(messageBrokerClient.subscribedSyncTopics, topic)
	return nil
}
