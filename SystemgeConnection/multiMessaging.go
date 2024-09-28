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
			connection.AsyncMessage(topic, payload)
		}(connection)
	}
	waitgroup.Wait()
}

func MultiSyncRequest(topic, payload string, connections ...SystemgeConnection) <-chan *Message.Message {
	responsesChannel := make(chan *Message.Message, len(connections))
	waitGroup := sync.WaitGroup{}
	for _, connection := range connections {
		responseChannel, _ := connection.SyncRequest(topic, payload)
		waitGroup.Add(1)
		go func(connection SystemgeConnection, responseChannel <-chan *Message.Message) {
			responsesChannel <- <-responseChannel
			waitGroup.Done()

		}(connection, responseChannel)
	}
	go func() {
		waitGroup.Wait()
		close(responsesChannel)
	}()
	return responsesChannel
}

func MultiSyncRequestBlocking(topic, payload string, connections ...SystemgeConnection) []*Message.Message {
	responses := make([]*Message.Message, len(connections))
	for i, connection := range connections {
		response, _ := connection.SyncRequestBlocking(topic, payload)
		responses[i] = response
	}
	return responses
}
