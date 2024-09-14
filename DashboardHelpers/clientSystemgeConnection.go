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

func NewSystemgeConnectionClient(name string, commands map[string]bool, status int, unprocessedMessages uint32) *SystemgeConnectionClient {
	return &SystemgeConnectionClient{
		Name:                    name,
		Commands:                commands,
		Status:                  status,
		UnprocessedMessageCount: unprocessedMessages,
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
	return &client, nil
}
