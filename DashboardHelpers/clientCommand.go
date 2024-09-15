package DashboardHelpers

import (
	"encoding/json"
	"time"
)

type CommandClient struct {
	Name     string                                `json:"name"`
	Commands map[string]bool                       `json:"commands"`
	Metrics  map[string]map[string][]*MetricsEntry `json:"metrics"` //periodically automatically updated by the server
}

func NewCommandClient(name string, commands map[string]bool, metrics map[string]map[time.Time]map[string]uint64) *CommandClient {
	metricsMap := map[string]map[string][]*MetricsEntry{}
	for key, value := range metrics {
		metricsMap[key] = map[string][]*MetricsEntry{}
		for key2, value2 := range value {
			metricsMap[key][key2] = append(metricsMap[key][key2], value2)
		}
	}
	return &CommandClient{
		Name:     name,
		Commands: commands,
		Metrics:  metricsMap,
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
