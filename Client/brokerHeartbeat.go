package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"time"
)

func (client *Client) heartbeatLoop(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		err := brokerConnection.send(Message.NewAsync("heartbeat", client.name, ""))
		if err != nil {
			client.logger.Log(Utilities.NewError("Failed to send heartbeat to message broker server \""+brokerConnection.resolution.Address+"\"", err).Error())
			return
		}
		sleepChannel := make(chan bool)
		go func() {
			time.Sleep(HEARTBEAT_INTERVAL)
			sleepChannel <- true
		}()
		select {
		case <-client.stopChannel:
			return
		case <-sleepChannel:
			continue
		}
	}
}
