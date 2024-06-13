package MessageBrokerServer

import (
	"Systemge/Message"
	"time"
)

func (client *Client) queueMessage(message *Message.Message) {
	client.messageQueue <- message
}

func (client *Client) dequeueMessage_Timeout(timeoutMs int) *Message.Message {
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
