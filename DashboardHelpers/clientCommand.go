package DashboardHelpers

import (
	"encoding/json"
)

type CommandClient struct {
	Name     string           `json:"name"`
	Commands map[string]bool  `json:"commands"`
	Metrics  DashboardMetrics `json:"metrics"`
}

func NewCommandClient(name string, commands map[string]bool, metrics DashboardMetrics) *CommandClient {
	return &CommandClient{
		Name:     name,
		Commands: commands,
		Metrics:  metrics,
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
		client.Metrics = DashboardMetrics{}
	}
	return &client, nil
}
