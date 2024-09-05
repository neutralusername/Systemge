package Commands

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

func NewCommandClient(name string, config *Config.CommandClient, command string, args ...string) (string, error) {
	connection, err := TcpConnection.EstablishConnection(config.TcpConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength)
	if err != nil {
		return "", Error.New("Failed to establish connection", err)
	}
	response, err := connection.SyncRequestBlocking("command", Helpers.JsonMarshal(commandStruct{
		Command: command,
		Args:    args,
	}))
	if err != nil {
		return "", Error.New("Failed to send command", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Error.New("Command requestFailed", errors.New((response.GetPayload())))
	}
	return response.GetPayload(), nil
}
