package Broker

import (
	"Systemge/Message"
	"time"
)

func (clientConnection *clientConnection) queueMessage(message *Message.Message) {
	clientConnection.messageQueue <- message
}

func (clientConnection *clientConnection) dequeueMessage_Timeout(timeoutMs int) *Message.Message {
	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop() // Ensure the timer is stopped to release resources
	select {
	case message := <-clientConnection.messageQueue:
		return message
	case <-timer.C:
		return nil
	}
}
