package Dashboard

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type client struct {
	Name           string            `json:"name"`
	Status         int               `json:"status"`
	Commands       map[string]bool   `json:"commands"`
	Metrics        map[string]uint64 `json:"metrics"`
	HasStatusFunc  bool              `json:"hasStatusFunc"`
	HasStartFunc   bool              `json:"hasStartFunc"`
	HasStopFunc    bool              `json:"hasStopFunc"`
	HasMetricsFunc bool              `json:"hasMetricsFunc"`

	connection *SystemgeConnection.SystemgeConnection
}

func unmarshalClient(data string) (*client, error) {
	var client client
	err := json.Unmarshal([]byte(data), &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (client *client) executeCommand(command string, args []string) (*Message.Message, error) {
	if client.connection == nil {
		return nil, Error.New("No connection available", nil)
	}
	if !client.Commands[command] {
		return nil, Error.New("Command not found", nil)
	}
	response, err := client.connection.SyncRequest(Message.TOPIC_EXECUTE_COMMAND, Helpers.JsonMarshal(&Command{
		Command: command,
		Args:    args,
	}))
	if err != nil {
		return nil, err
	}
	return response, nil
}
