package Client

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

func (client *Client) heartbeatLoop(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		err := brokerConnection.send(Message.NewAsync("heartbeat", client.config.Name, ""))
		if err != nil {
			client.logger.Log(Error.New("Failed to send heartbeat to message broker server \""+brokerConnection.resolution.GetAddress()+"\"", err).Error())
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
