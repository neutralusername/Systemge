package Commands

import (
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

func NewCommandClient(name string, config *Config.CommandClient, command string, args ...string) (string, error) {
	connection, err := TcpConnection.EstablishConnection(config.TcpConnectionConfig, config.TcpClientConfig, name, int(config.MaxServerNameLength))
	if err != nil {
		return "", Error.New("Failed to establish connection", err)
	}
	commandStruct := commandStruct{
		Command: command,
		Args:    args,
	}
	response, err := connection.SyncRequestBlocking("command", Helpers.JsonMarshal(commandStruct))
	if err != nil {
		return "", Error.New("Failed to send command", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Error.New("Command failed", Error.New(response.GetPayload(), nil))
	}
	return response.GetPayload(), nil
}
