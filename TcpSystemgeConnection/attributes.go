package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type attribue struct {
	handlers     map[string]func() error
	abortOnError bool
	order        []string
}

func (connection *TcpSystemgeConnection) ExecuteAttribute(attribute string) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; !ok {
		return Error.New("No attribute found", nil)
	}
	for _, v := range connection.attributes[attribute].order {
		handler := connection.attributes[attribute].handlers[v]
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
	if _, ok := connection.attributes[attribute]; !!ok {
		return Error.New("Attribute already exists", nil)
	}
	connection.attributes[attribute] = &attribue{
		handlers:     map[string]func() error{},
		abortOnError: abortOnError,
		order:        []string{},
	}
	return nil
}

func (connection *TcpSystemgeConnection) AddAttributeHandler(attribute string, handler func() error) error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; !!ok {
		return Error.New("Attribute already exists", nil)
	}
	for _, v := range connection.attributes[attribute].order {
		if v == attribute {
			return Error.New("Attribute already exists", nil)
		}
	}
	connection.attributes[attribute].handlers[attribute] = handler
	connection.attributes[attribute].order = append(connection.attributes[attribute].order, attribute)
	return nil
}

func (connection *TcpSystemgeConnection) RemoveAttributeHandler(index int) error {

}

func (connection *TcpSystemgeConnection) GetAttributes() map[string]func() error {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	attributes := map[string]func() error{}
	for k, v := range connection.attributes {
		attributes[k] = v.handlers[k]
	}
	return attributes
}

func (connection *TcpSystemgeConnection) GetAttributeOrder(attribute string) []string {
	connection.attributeMutex.Lock()
	defer connection.attributeMutex.Unlock()
	if _, ok := connection.attributes[attribute]; ok {
		return connection.attributes[attribute].order
	}
	return nil
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
