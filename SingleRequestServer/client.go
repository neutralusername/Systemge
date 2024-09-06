package SingleRequestServer

import (
	"encoding/json"
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/TcpConnection"
)

type commandStruct struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func UnmarshalCommandStruct(data string) *commandStruct {
	var command commandStruct
	err := json.Unmarshal([]byte(data), &command)
	if err != nil {
		return nil
	}
	return &command
}

func RemoteCommandRequest(name string, config *Config.RemoteCommand, command string, args ...string) (string, error) {
	connection, err := TcpConnection.EstablishConnection(config.TcpConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength)
	if err != nil {
		return "", Error.New("Failed to establish connection", err)
	}
	err = connection.AsyncMessage("command", Helpers.JsonMarshal(commandStruct{
		Command: command,
		Args:    args,
	}))
	if err != nil {
		return "", Error.New("Failed to send command", err)
	}
	message, err := connection.GetNextMessage()
	if err != nil {
		return "", Error.New("Failed to get response", err)
	}
	if message.GetTopic() != Message.TOPIC_SUCCESS {
		return "", Error.New("Command failed", errors.New(message.GetPayload()))
	}
	connection.SyncResponse(message, true, "")
	connection.Close()
	return message.GetPayload(), nil
}

func RemoteAsyncMessage(name string, config *Config.RemoteCommand, topic string, payload string) error {
	connection, err := TcpConnection.EstablishConnection(config.TcpConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength)
	if err != nil {
		return Error.New("Failed to establish connection", err)
	}
	err = connection.AsyncMessage("async", string(Message.NewAsync(topic, payload).Serialize()))
	if err != nil {
		return Error.New("Failed to send message", err)
	}
	message, err := connection.GetNextMessage()
	if err != nil {
		return Error.New("Failed to get response", err)
	}
	if message.GetTopic() != Message.TOPIC_SUCCESS {
		return Error.New("Message failed", errors.New(message.GetPayload()))
	}
	connection.SyncResponse(message, true, "")
	connection.Close()
	return nil
}

func RemoteSyncRequest(name string, config *Config.RemoteCommand, topic string, payload string) (string, error) {
	connection, err := TcpConnection.EstablishConnection(config.TcpConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength)
	if err != nil {
		return "", Error.New("Failed to establish connection", err)
	}
	err = connection.AsyncMessage("sync", string(Message.NewAsync(topic, payload).Serialize()))
	if err != nil {
		return "", Error.New("Failed to send request", err)
	}
	message, err := connection.GetNextMessage()
	if err != nil {
		return "", Error.New("Failed to get response", err)
	}
	if message.GetTopic() != Message.TOPIC_SUCCESS {
		return "", Error.New("Request failed", errors.New(message.GetPayload()))
	}
	connection.SyncResponse(message, true, "")
	connection.Close()
	return message.GetPayload(), nil
}
