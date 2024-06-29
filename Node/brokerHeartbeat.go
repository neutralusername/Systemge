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
			node.logger.Log(Error.New("Failed to send heartbeat to broker \""+brokerConnection.resolution.GetAddress()+"\"", err).Error())
			return
		}
		sleepChannel := make(chan bool)
		go func() {
			time.Sleep(HEARTBEAT_INTERVAL)
			sleepChannel <- true
		}()
		select {
		case <-node.stopChannel:
			return
		case <-sleepChannel:
			continue
		}
	}
}
