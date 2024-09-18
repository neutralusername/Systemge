package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type attributeStruct struct {
	responseChannel chan *Message.Message
	abortChannel    chan bool
}

func (connection *TcpSystemgeConnection) AbortSyncRequest(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if syncRequestStruct, ok := connection.attributes[attribute]; ok {
		close(syncRequestStruct.abortChannel)
		delete(connection.attributes, attribute)
		return nil
	}
	return Error.New("No response channel found", nil)
}

// returns a slice of syncTokens of open sync requests
func (connection *TcpSystemgeConnection) GetOpenSyncRequests() []string {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	syncTokens := make([]string, 0, len(connection.attributes))
	for k := range connection.attributes {
		syncTokens = append(syncTokens, k)
	}
	return syncTokens
}

func (connection *TcpSystemgeConnection) initResponseChannel() (string, *attributeStruct) {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	syncToken := connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	for _, ok := connection.attributes[syncToken]; ok; {
		syncToken = connection.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	}
	connection.attributes[syncToken] = &attributeStruct{
		responseChannel: make(chan *Message.Message, 1),
		abortChannel:    make(chan bool),
	}
	return syncToken, connection.attributes[syncToken]
}

func (connection *TcpSystemgeConnection) addSyncResponse(message *Message.Message) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if syncRequestStruct, ok := connection.attributes[message.GetSyncToken()]; ok {
		syncRequestStruct.responseChannel <- message
		close(syncRequestStruct.responseChannel)
		delete(connection.attributes, message.GetSyncToken())
		return nil
	}
	return Error.New("No response channel found", nil)
}

func (connection *TcpSystemgeConnection) removeSyncRequest(syncToken string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[syncToken]; ok {
		delete(connection.attributes, syncToken)
		return nil
	}
	return Error.New("No response channel found", nil)
}
