package DashboardHelpers

import (
	"encoding/json"
)

type CommandClient struct {
	Name     string                                `json:"name"`
	Commands map[string]bool                       `json:"commands"`
	Metrics  map[string]map[string][]*MetricsEntry `json:"metrics"` //periodically automatically updated by the server
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
	if client.Commands == nil {
		client.Commands = map[string]bool{}
	}
	if client.Metrics == nil {
		client.Metrics = map[string]map[string][]*MetricsEntry{}
	}
	return &client, nil
}
