package Broker

import (
	"Systemge/Message"
	"time"
)

func (nodeConnection *nodeConnection) queueMessage(message *Message.Message) {
	nodeConnection.messageQueue <- message
}

func (nodeConnection *nodeConnection) dequeueMessage_Timeout(timeoutMs int) *Message.Message {
	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop() // Ensure the timer is stopped to release resources
	select {
	case message := <-nodeConnection.messageQueue:
		return message
	case <-timer.C:
		return nil
	}
}
