package DashboardHelpers

import (
	"encoding/json"
)

type CustomServiceClient struct {
	Name     string           `json:"name"`
	Commands map[string]bool  `json:"commands"`
	Status   int              `json:"status"` //periodically automatically updated by the server
	Metrics  DashboardMetrics `json:"metrics"`
}

func NewCustomServiceClient(name string, commands map[string]bool, status int, metrics DashboardMetrics) *CustomServiceClient {
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
	if client.Commands == nil {
		client.Commands = map[string]bool{}
	}
	if client.Metrics == nil {
		client.Metrics = DashboardMetrics{}
	}
	return &client, nil
}
