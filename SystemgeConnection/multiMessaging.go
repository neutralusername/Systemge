package SystemgeConnection

import (
	"sync"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func MultiAsyncMessage(topic, payload string, connections ...SystemgeConnection) <-chan error {
	errorChannel := make(chan error, len(connections))
	for _, connection := range connections {
		err := connection.AsyncMessage(topic, payload)
		if err != nil {
			errorChannel <- Event.New("Failed to send AsyncMessage to \""+connection.GetName()+"\"", err)
		}
	}
	close(errorChannel)
	return errorChannel
}

func MultiSyncRequest(topic, payload string, connections ...SystemgeConnection) (<-chan *Message.Message, <-chan error) {
	responsesChannel := make(chan *Message.Message, len(connections))
	errorChannel := make(chan error, len(connections))
	waitGroup := sync.WaitGroup{}
	for _, connection := range connections {
		responseChannel, err := connection.SyncRequest(topic, payload)
		if err != nil {
			errorChannel <- Event.New("Failed to send SyncRequest to \""+connection.GetName()+"\"", err)
		} else {
			waitGroup.Add(1)
			go func(connection SystemgeConnection, responseChannel <-chan *Message.Message) {
				responseMessage := <-responseChannel
				if responseMessage == nil {
					errorChannel <- Event.New("Failed to receive response from \""+connection.GetName()+"\"", nil)
				} else {
					responsesChannel <- responseMessage
				}
				waitGroup.Done()

			}(connection, responseChannel)
		}
	}
	go func() {
		waitGroup.Wait()
		close(responsesChannel)
		close(errorChannel)
	}()
	return responsesChannel, errorChannel
}
