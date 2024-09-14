package DashboardHelpers

import (
	"encoding/json"
)

type CommandClient struct {
	Name     string                              `json:"name"`
	Commands map[string]bool                     `json:"commands"`
	Metrics  map[string]map[string]*MetricsEntry `json:"metrics"`
}

func NewCommandClient(name string, commands map[string]bool) *CommandClient {
	return &CommandClient{
		Name:     name,
		Commands: commands,
	}
}

func (client *CommandClient) Marshal() []byte {
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
