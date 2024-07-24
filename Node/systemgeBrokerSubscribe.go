package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Tools"
	"time"
)

func (node *Node) subscribeLoop(topic string, maxSubscribeAttempts uint64) error {
	subscribeAttempts := uint64(0)
	systemge_ := node.systemge
	for {
		systemge := node.systemge
		if systemge == nil {
			return Error.New("systemge not initialized", nil)
		}
		if systemge != systemge_ {
			return Error.New("systemge changed", nil)
		}
		if subscribeAttempts >= maxSubscribeAttempts && maxSubscribeAttempts > 0 {
			return Error.New("Reached maximum subscribe attempts", nil)
		}
		subscribeAttempts++
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Attempting subscription to topic \""+topic+"\"", nil).Error())
		}
		brokerConnection, err := node.getBrokerConnectionForTopic(topic, false)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to subscribe to topic \""+topic+"\"", err).Error())
			}
			time.Sleep(time.Duration(systemge.application.GetSystemgeComponentConfig().BrokerSubscribeDelayMs) * time.Millisecond)
			continue
		}
		go node.handleBrokerConnectionMessages(brokerConnection)
		err = systemge.subscribeTopic(node.GetName(), brokerConnection, topic)
		if err != nil {
			brokerConnection.mutex.Lock()
			subscribedTopicsCount := len(brokerConnection.subscribedTopics)
			topicResolutionsCount := len(brokerConnection.topicResolutions)
			if subscribedTopicsCount == 0 && topicResolutionsCount == 0 {
				brokerConnection.close()
			}
			brokerConnection.mutex.Unlock()
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to subscribe to topic \""+topic+"\"", err).Error())
			}
			time.Sleep(time.Duration(systemge.application.GetSystemgeComponentConfig().BrokerSubscribeDelayMs) * time.Millisecond)
			continue
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Added subscribed topic for broker \""+brokerConnection.endpoint.Address+"\" for topic \""+topic+"\"", nil).Error())
		}
		return nil
	}
}

func (systemge *systemgeComponent) subscribeTopic(nodeName string, brokerConnection *brokerConnection, topic string) error {
	message := Message.NewSync("subscribe", nodeName, topic, Tools.RandomString(10, Tools.ALPHA_NUMERIC))
	responseChannel, err := systemge.addMessageWaitingForResponse(message)
	if err != nil {
		return Error.New("failed to add message waiting for response", err)
	}
	bytesSent, err := brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message.Serialize())
	if err != nil {
		systemge.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to broker", err)
	}
	systemge.bytesSentCounter.Add(bytesSent)
	response, err := systemge.receiveSyncResponse(message, responseChannel)
	if err != nil {
		return Error.New("Error sending subscription sync request", err)
	}
	if response.GetTopic() != "subscribed" {
		return Error.New("Invalid response topic \""+response.GetTopic()+"\"", nil)
	}
	err = brokerConnection.addSubscribedTopic(topic)
	if err != nil {
		return Error.New("Failed to add subscribed topic for broker \""+brokerConnection.endpoint.Address+"\" for topic \""+topic+"\"", err)
	}
	return nil
}
