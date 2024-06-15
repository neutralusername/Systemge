package Broker

import (
	"Systemge/Message"
	"time"
)

func (client *ClientConnection) queueMessage(message *Message.Message) {
	client.messageQueue <- message
}

func (client *ClientConnection) dequeueMessage_Timeout(timeoutMs int) *Message.Message {
	timeoutChannel := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond * time.Duration(timeoutMs))
		timeoutChannel <- true
	}()
	select {
	case message := <-client.messageQueue:
		return message
	case <-timeoutChannel:
		return nil
	}
}
