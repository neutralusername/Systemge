package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type attribue struct {
	abortOnError bool
	order        []func() error
}

func (connection *TcpSystemgeConnection) ExecuteAttribute(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; !ok {
		return Error.New("No attribute found", nil)
	}
	for _, handler := range connection.attributes[attribute].order {
		err := handler()
		if err != nil && connection.attributes[attribute].abortOnError {
			return err
		}
	}
	return nil
}

func (connection *TcpSystemgeConnection) NewAttribute(attribute string, abortOnError bool) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		return Error.New("Attribute already exists", nil)
	}
	connection.attributes[attribute] = &attribue{
		abortOnError: abortOnError,
		order:        []func() error{},
	}
	return nil
}

func (connection *TcpSystemgeConnection) RemoveAttribute(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		delete(connection.attributes, attribute)
		return nil
	}
	return Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) AppendAttributeHandler(attribute string, handler func() error) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; !ok {
		return Error.New("No attribute found", nil)
	}
	connection.attributes[attribute].order = append(connection.attributes[attribute].order, handler)
	return nil
}

func (connection *TcpSystemgeConnection) PopAttributeHandler(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		connection.attributes[attribute].order = connection.attributes[attribute].order[:len(connection.attributes[attribute].order)-1]
		return nil
	}
	return Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) GetAttributeHandlerCount(attribute string) (int, error) {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		return len(connection.attributes[attribute].order), nil
	}
	return -1, Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) ClearAttributeHandlers(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		connection.attributes[attribute].order = []func() error{}
		return nil
	}
	return Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) SetAttributeHandlers(attribute string, handlers []func() error) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		connection.attributes[attribute].order = handlers
		return nil
	}
	return Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) InsertAttributeHandler(attribute string, index int, handler func() error) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		if index < 0 || index > len(connection.attributes[attribute].order) {
			return Error.New("Index out of bounds", nil)
		}
		connection.attributes[attribute].order = append(connection.attributes[attribute].order[:index], append([]func() error{handler}, connection.attributes[attribute].order[index:]...)...)
		return nil
	}
	return Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) RemoveAttributeHandler(attribute string, index int) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		if index < 0 || index >= len(connection.attributes[attribute].order) {
			return Error.New("Index out of bounds", nil)
		}
		connection.attributes[attribute].order = append(connection.attributes[attribute].order[:index], connection.attributes[attribute].order[index+1:]...)
		return nil
	}
	return Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) GetAttributeHandler(attribute string, index int) (func() error, error) {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		if index < 0 || index >= len(connection.attributes[attribute].order) {
			return nil, Error.New("Index out of bounds", nil)
		}
		return connection.attributes[attribute].order[index], nil
	}
	return nil, Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) GetAttributeHandlers(attribute string) ([]func() error, error) {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		return connection.attributes[attribute].order, nil
	}
	return nil, Error.New("No attribute found", nil)
}

func (connection *TcpSystemgeConnection) SwapAttributeOrder(attribute string, index1, index2 int) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		if index1 < 0 || index1 >= len(connection.attributes[attribute].order) || index2 < 0 || index2 >= len(connection.attributes[attribute].order) {
			return Error.New("Index out of bounds", nil)
		}
		connection.attributes[attribute].order[index1], connection.attributes[attribute].order[index2] = connection.attributes[attribute].order[index2], connection.attributes[attribute].order[index1]
		return nil
	}
	return Error.New("No attribute found", nil)
}

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
