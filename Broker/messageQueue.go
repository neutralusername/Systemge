package Broker

import (
	"Systemge/Message"
	"time"
)

func (clientConnection *clientConnection) queueMessage(message *Message.Message) {
	clientConnection.messageQueue <- message
}

func (clientConnection *clientConnection) dequeueMessage_Timeout(timeoutMs int) *Message.Message {
	timeoutChannel := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond * time.Duration(timeoutMs))
		timeoutChannel <- true
	}()
	select {
	case message := <-clientConnection.messageQueue:
		return message
	case <-timeoutChannel:
		return nil
	}
}
