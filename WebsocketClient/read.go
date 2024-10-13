package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) Read() ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	return client.read()
}

func (client *WebsocketClient) read() ([]byte, error) {
	_, messageBytes, err := client.websocketConn.ReadMessage()
	if err != nil {
		if isWebsocketConnClosedErr(err) {
			client.Close()
		}
		return nil, err
	}
	client.BytesReceived.Add(uint64(len(messageBytes)))
	client.MessagesReceived.Add(1)
	return messageBytes, nil
}

// can be used to cancel an ongoing read operation
func (client *WebsocketClient) SetReadDeadline(timeoutMs uint64) {
	client.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}

func (client *WebsocketClient) StartReadHandler(receptionHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.receptionHandler = receptionHandler
	go client.readRoutine()

	return nil
}

func (client *WebsocketClient) StopReadHandler() error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler == nil {
		return errors.New("receptionHandler is not running")
	}

	// close(readRoutineChannel)
	client.websocketConn.SetReadDeadline(time.Now())
	// wg.Wait()
	client.receptionHandler = nil
	return nil
}

func (client *WebsocketClient) readRoutine() {
	for {
		bytes, err := client.read()
		if err != nil {
			continue
		}

		client.receptionHandler(bytes, client)
	}
}
