package Node

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
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
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Attempting subscription to topic \""+topic+"\" (attempt "+Helpers.Uint64ToString(subscribeAttempts+1)+"). Getting broker connection", nil).Error())
		}
		brokerConnection, err := node.getBrokerConnectionForTopic(topic, false)
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to get broker connection for topic \""+topic+"\" (attempt "+Helpers.Uint64ToString(subscribeAttempts)+")", err).Error())
			}
			time.Sleep(time.Duration(systemge.application.GetSystemgeComponentConfig().BrokerSubscribeDelayMs) * time.Millisecond)
			continue
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Got broker connection \""+brokerConnection.endpoint.Address+"\" for topic \""+topic+"\" (attempt "+Helpers.Uint64ToString(subscribeAttempts)+"). Subscribing to topic", nil).Error())
		}
		err = systemge.subscribeTopic(node.GetName(), brokerConnection, topic)
		if err != nil {
			if brokerConnection.closeIfNoTopics() {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Closed broker connection \""+brokerConnection.endpoint.Address+"\" due to no topics (attempt "+Helpers.Uint64ToString(subscribeAttempts)+")", nil).Error())
				}
			}
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to subscribe to topic \""+topic+"\" from broker connection \""+brokerConnection.endpoint.Address+"\" (attempt "+Helpers.Uint64ToString(subscribeAttempts)+")", err).Error())
			}
			time.Sleep(time.Duration(systemge.application.GetSystemgeComponentConfig().BrokerSubscribeDelayMs) * time.Millisecond)
			continue
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Subscribed to topic \""+topic+"\" from broker connection \""+brokerConnection.endpoint.Address+"\" (attempt "+Helpers.Uint64ToString(subscribeAttempts)+")", nil).Error())
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
		return Error.New("Failed to send subscription sync request", err)
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
		return Error.New("Failed to add subscribed topic", err)
	}
	return nil
}
