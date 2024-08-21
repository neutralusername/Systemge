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

func (client *client) executeCommand(command string, args []string) (string, error) {
	if !client.Commands[command] {
		return "", Error.New("Command \""+command+"\" not found", nil)
	}
	response, err := client.connection.SyncRequest(Message.TOPIC_EXECUTE_COMMAND, Helpers.JsonMarshal(&Command{
		Command: command,
		Args:    args,
	}))
	if err != nil {
		return "", Error.New("Failed to send command \""+command+"\" to client \""+client.Name+"\"", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Error.New(response.GetPayload(), nil)
	}
	return response.GetPayload(), nil
}
