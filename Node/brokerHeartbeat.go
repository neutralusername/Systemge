package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

func (node *Node) heartbeatLoop(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		brokerConnection.mutex.Lock()
		topicResolutionCount := len(brokerConnection.topicResolutions)
		subscribedTopicCount := len(brokerConnection.subscribedTopics)
		brokerConnection.mutex.Unlock()
		if topicResolutionCount+subscribedTopicCount != 0 {
			err := brokerConnection.send(Message.NewAsync("heartbeat", node.config.Name, ""))
			if err != nil {
				node.logger.Log(Error.New("Failed to send heartbeat to broker \""+brokerConnection.endpoint.GetAddress()+"\"", err).Error())
				return
			}
		}
		sleepTimeout := time.NewTimer(time.Duration(node.config.BrokerHeartbeatIntervalMs) * time.Millisecond)
		select {
		case <-node.stopChannel:
			sleepTimeout.Stop()
			return
		case <-sleepTimeout.C:
			continue
		}
	}
}
