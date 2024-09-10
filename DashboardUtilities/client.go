package DashboardUtilities

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type Client struct {
	Name           string            `json:"name"`
	Status         int               `json:"status"`
	Commands       map[string]bool   `json:"commands"`
	Metrics        map[string]uint64 `json:"metrics"`
	HasStatusFunc  bool              `json:"hasStatusFunc"`
	HasStartFunc   bool              `json:"hasStartFunc"`
	HasStopFunc    bool              `json:"hasStopFunc"`
	HasMetricsFunc bool              `json:"hasMetricsFunc"`

	Connection           SystemgeConnection.SystemgeConnection
	visitingWebsocketIds map[string]bool
}

func UnmarshalClient(data string) (*Client, error) {
	var client Client
	err := json.Unmarshal([]byte(data), &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (client *Client) ExecuteCommand(command string, args []string) (string, error) {
	if !client.Commands[command] {
		return "", Error.New("Command \""+command+"\" not found", nil)
	}
	response, err := client.Connection.SyncRequestBlocking(Message.TOPIC_EXECUTE_COMMAND, Helpers.JsonMarshal(&Command{
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
