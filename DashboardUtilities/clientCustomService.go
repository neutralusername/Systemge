package DashboardUtilities

import (
	"encoding/json"
)

type CustomServiceClient struct {
	Name     string            `json:"name"`
	Commands []string          `json:"commands"`
	Status   int               `json:"status"`
	Metrics  map[string]uint64 `json:"metrics"`
}

func NewCustomServiceClient(name string, commands []string, status int, metrics map[string]uint64) *CustomServiceClient {
	return &CustomServiceClient{
		Name:     name,
		Commands: commands,
		Status:   status,
		Metrics:  metrics,
	}
}
func (client *CustomServiceClient) Marshal() []byte {
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

/* server code

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
*/
