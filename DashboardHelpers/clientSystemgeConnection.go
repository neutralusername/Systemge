package DashboardHelpers

import "encoding/json"

type SystemgeConnectionClient struct {
	Name                    string                                `json:"name"`
	Commands                map[string]bool                       `json:"commands"`
	Status                  int                                   `json:"status"`                  //periodically automatically updated by the server
	IsProcessingLoopRunning bool                                  `json:"isProcessingLoopRunning"` //periodically automatically updated by the server
	UnprocessedMessageCount uint32                                `json:"unprocessedMessageCount"` //periodically automatically updated by the server
	Metrics                 map[string]map[string][]*MetricsEntry `json:"metrics"`                 //periodically automatically updated by the server
}

func NewSystemgeConnectionClient(name string, commands map[string]bool, status int, unprocessedMessages uint32, metrics map[string]map[string]*MetricsEntry) *SystemgeConnectionClient {
	metricsMap := map[string]map[string][]*MetricsEntry{}
	for key, value := range metrics {
		metricsMap[key] = map[string][]*MetricsEntry{}
		for key2, value2 := range value {
			metricsMap[key][key2] = append(metricsMap[key][key2], value2)
		}
	}
	return &SystemgeConnectionClient{
		Name:                    name,
		Commands:                commands,
		Status:                  status,
		UnprocessedMessageCount: unprocessedMessages,
		Metrics:                 metricsMap,
	}
}
func (client *SystemgeConnectionClient) Marshal() []byte {
	bytes, err := json.Marshal(client)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalSystemgeConnectionClient(bytes []byte) (*SystemgeConnectionClient, error) {
	var client SystemgeConnectionClient
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
