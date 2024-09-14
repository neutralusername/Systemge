package DashboardHelpers

import (
	"encoding/json"
)

type CustomServiceClient struct {
	Name     string                                `json:"name"`
	Commands map[string]bool                       `json:"commands"`
	Status   int                                   `json:"status"` //periodically automatically updated by the server
	Metrics  map[string]map[string][]*MetricsEntry `json:"metrics"`
}

func NewCustomServiceClient(name string, commands map[string]bool, status int) *CustomServiceClient {
	return &CustomServiceClient{
		Name:     name,
		Commands: commands,
		Status:   status,
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
