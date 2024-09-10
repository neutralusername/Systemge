package DashboardUtilities

import (
	"encoding/json"
)

type CustomServiceClient struct {
	Name     string            `json:"name"`
	Status   int               `json:"status"`
	Commands []string          `json:"commands"`
	Metrics  map[string]uint64 `json:"metrics"`
}

func (client *CustomServiceClient) GetClientType() int {
	return CLIENT_CUSTOM_SERVICE
}

func MarshalCustomClient(name string, commands []string, status int, metrics map[string]uint64) []byte {
	client := CustomServiceClient{
		Name:     name,
		Status:   status,
		Commands: commands,
		Metrics:  metrics,
	}
	bytes, err := json.Marshal(client)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalCustomClient(bytes []byte) (*CustomServiceClient, error) {
	var client CustomServiceClient
	err := json.Unmarshal(bytes, &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

/* func (client *Client) ExecuteCommand(command string, args []string) (string, error) {
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
*/
