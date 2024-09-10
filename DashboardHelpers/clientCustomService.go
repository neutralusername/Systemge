package DashboardHelpers

import (
	"encoding/json"
)

type CustomServiceClient struct {
	Name     string            `json:"name"`
	Commands map[string]bool   `json:"commands"`
	Status   int               `json:"status"`
	Metrics  map[string]uint64 `json:"metrics"`
}

func NewCustomServiceClient(name string, commands map[string]bool, status int, metrics map[string]uint64) *CustomServiceClient {
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
