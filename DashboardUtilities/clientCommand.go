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

func MarshalCommandClient(name string, commands []string) []byte {
	client := CommandClient{
		Name:     name,
		Commands: commands,
	}
	bytes, err := json.Marshal(client)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalCommandClient(bytes []byte) (*CommandClient, error) {
	var client CommandClient
	err := json.Unmarshal(bytes, &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
