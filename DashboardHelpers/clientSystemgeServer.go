package DashboardHelpers

import (
	"encoding/json"
)

type SystemgeServerClient struct {
	Name                       string                              `json:"name"`
	Commands                   map[string]bool                     `json:"commands"`
	Status                     int                                 `json:"status"`                     //periodically automatically updated by the server
	Metrics                    DashboardMetrics                    `json:"metrics"`                    //periodically automatically updated by the server
	SystemgeConnectionChildren map[string]*SystemgeConnectionChild `json:"systemgeConnectionChildren"` //periodically automatically updated by the server
}

type SystemgeConnectionChild struct {
	Name                    string `json:"name"`
	IsProcessingLoopRunning bool   `json:"isProcessingLoopRunning"`
	UnprocessedMessageCount uint32 `json:"unprocessedMessageCount"`
}

func NewSystemgeConnectionChild(name string, isProcessingLoopRunning bool, unprocessedMessages uint32) *SystemgeConnectionChild {
	return &SystemgeConnectionChild{
		Name:                    name,
		IsProcessingLoopRunning: isProcessingLoopRunning,
		UnprocessedMessageCount: unprocessedMessages,
	}
}

func NewSystemgeServerClient(name string, commands map[string]bool, status int, metrics DashboardMetrics, systemgeConnections map[string]*SystemgeConnectionChild) *SystemgeServerClient {
	return &SystemgeServerClient{
		Name:                       name,
		Commands:                   commands,
		Status:                     status,
		Metrics:                    metrics,
		SystemgeConnectionChildren: systemgeConnections,
	}
}
func (client *SystemgeServerClient) Marshal() []byte {
	bytes, err := json.Marshal(client)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalSystemgeServerClient(bytes []byte) (*SystemgeServerClient, error) {
	var client SystemgeServerClient
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
