package DashboardUtilities

import (
	"encoding/json"
)

type CommandClient struct {
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
}

func (client *CommandClient) GetClientType() int {
	return CLIENT_COMMAND
}

func UnmarshalCommandClient(data string) (*CommandClient, error) {
	var client CommandClient
	err := json.Unmarshal([]byte(data), &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
