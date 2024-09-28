package SystemgeConnection

import (
	"sync"

	"github.com/neutralusername/Systemge/Message"
)

func MultiAsyncMessage(topic, payload string, connections ...SystemgeConnection) {
	waitgroup := sync.WaitGroup{}
	for _, connection := range connections {
		waitgroup.Add(1)
		go func(connection SystemgeConnection) {
			defer waitgroup.Done()
			connection.AsyncMessage(topic, payload)
		}(connection)
	}
	waitgroup.Wait()
}

func MultiSyncRequest(topic, payload string, connections ...SystemgeConnection) <-chan *Message.Message {
	responsesChannel := make(chan *Message.Message, len(connections))
	responseWaitGroup := sync.WaitGroup{}
	sendWaitGroup := sync.WaitGroup{}
	for _, connection := range connections {
		sendWaitGroup.Add(1)
		go func(connection SystemgeConnection) {
			defer sendWaitGroup.Done()
			responseChannel, err := connection.SyncRequest(topic, payload)
			if err != nil {
				return
			}
			responseWaitGroup.Add(1)
			go func(connection SystemgeConnection, responseChannel <-chan *Message.Message) {
				responsesChannel <- <-responseChannel
				responseWaitGroup.Done()
			}(connection, responseChannel)
		}(connection)
	}
	go func() {
		responseWaitGroup.Wait()
		close(responsesChannel)
	}()
	return responsesChannel
}

func MultiSyncRequestBlocking(topic, payload string, connections ...SystemgeConnection) []*Message.Message {
	responses := []*Message.Message{}
	waitgroup := sync.WaitGroup{}
	for _, connection := range connections {
		waitgroup.Add(1)
		go func(connection SystemgeConnection) {
			defer waitgroup.Done()
			response, err := connection.SyncRequestBlocking(topic, payload)
			if err != nil {
				return
			}
			responses = append(responses, response)
		}(connection)
	}
	return responses
}
