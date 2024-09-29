package SingleRequestServer

import (
	"encoding/json"
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/TcpSystemgeConnect"
)

type commandStruct struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func unmarshalCommandStruct(data string) *commandStruct {
	var command commandStruct
	err := json.Unmarshal([]byte(data), &command)
	if err != nil {
		return nil
	}
	return &command
}

func Command(name string, config *Config.SingleRequestClient, command string, args ...string) (string, error) {
	connection, err := TcpSystemgeConnect.EstablishConnection(config.TcpSystemgeConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength, nil)
	if err != nil {
		return "", err
	}
	err = connection.AsyncMessage("command", Helpers.JsonMarshal(commandStruct{
		Command: command,
		Args:    args,
	}))
	if err != nil {
		return "", err
	}
	message, err := connection.RetrieveNextMessage()
	if err != nil {
		return "", err
	}
	connection.SyncResponse(message, true, "")
	if message.GetTopic() != Message.TOPIC_SUCCESS {
		return "", errors.New(message.GetPayload())
	}
	connection.Close()
	return message.GetPayload(), nil
}

func AsyncMessage(name string, config *Config.SingleRequestClient, topic string, payload string) error {
	connection, err := TcpSystemgeConnect.EstablishConnection(config.TcpSystemgeConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength, nil)
	if err != nil {
		return err
	}
	err = connection.AsyncMessage("async", string(Message.NewAsync(topic, payload).Serialize()))
	if err != nil {
		return err
	}
	connection.Close()
	return nil
}

func SyncRequest(name string, config *Config.SingleRequestClient, topic string, payload string) (*Message.Message, error) {
	connection, err := TcpSystemgeConnect.EstablishConnection(config.TcpSystemgeConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength, nil)
	if err != nil {
		return nil, err
	}
	err = connection.AsyncMessage("sync", string(Message.NewAsync(topic, payload).Serialize()))
	if err != nil {
		return nil, err
	}
	message, err := connection.RetrieveNextMessage()
	if err != nil {
		return nil, err
	}
	connection.SyncResponse(message, true, "")
	connection.Close()
	return message, nil
}
