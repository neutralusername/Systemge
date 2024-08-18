package SystemgeConnection

import (
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

func MultiAsyncMessage(topic, payload string, connections ...*SystemgeConnection) <-chan error {
	errorChannel := make(chan error, len(connections))
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(connections))
	for _, connection := range connections {
		go func(connection *SystemgeConnection) {
			err := connection.AsyncMessage(topic, payload)
			if err != nil {
				errorChannel <- Error.New("Error in AsyncMessage for \""+connection.GetName()+"\"", err)
			}
			waitGroup.Done()
		}(connection)
	}
	go func() {
		waitGroup.Wait()
		close(errorChannel)
	}()
	return errorChannel
}

func MultiSyncRequest(topic, payload string, connections ...*SystemgeConnection) (<-chan *Message.Message, <-chan error) {
	responseChannel := make(chan *Message.Message, len(connections))
	errorChannel := make(chan error, len(connections))
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(connections))
	for _, connection := range connections {
		go func(connection *SystemgeConnection) {
			response, err := connection.SyncRequest(topic, payload)
			if err != nil {
				errorChannel <- Error.New("Error in SyncRequest for \""+connection.GetName()+"\"", err)
			} else {
				responseChannel <- response
			}
			waitGroup.Done()
		}(connection)
	}
	go func() {
		waitGroup.Wait()
		close(responseChannel)
		close(errorChannel)
	}()
	return responseChannel, errorChannel
}
