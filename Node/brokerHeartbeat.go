package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

func (node *Node) heartbeatLoop(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		err := brokerConnection.send(Message.NewAsync("heartbeat", node.config.Name, ""))
		if err != nil {
			node.logger.Log(Error.New("Failed to send heartbeat to broker \""+brokerConnection.endpoint.GetAddress()+"\"", err).Error())
			return
		}
		sleepTimeout := time.NewTimer(time.Duration(node.config.HeartbeatIntervalMs) * time.Millisecond)
		select {
		case <-node.stopChannel:
			sleepTimeout.Stop()
			return
		case <-sleepTimeout.C:
			continue
		}
	}
}
