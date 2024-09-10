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

func UnmarshalCommandClient(bytes []byte) (*CommandClient, error) {
	var client CommandClient
	err := json.Unmarshal(bytes, &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
