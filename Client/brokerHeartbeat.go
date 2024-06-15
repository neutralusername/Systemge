package Client

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
)

func (client *Client) heartbeatLoop(serverConnection *brokerConnection) {
	for serverConnection.netConn != nil {
		err := serverConnection.send(Message.NewAsync("heartbeat", client.name, ""))
		if err != nil {
			client.logger.Log(Error.New("Failed to send heartbeat to message broker server \""+serverConnection.resolution.Address+"\"", err).Error())
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
