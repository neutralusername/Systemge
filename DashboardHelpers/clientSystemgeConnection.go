package DashboardHelpers

import "encoding/json"

type SystemgeConnectionClient struct {
	Name                string            `json:"name"`
	Commands            []string          `json:"commands"`
	Status              int               `json:"status"`
	Metrics             map[string]uint64 `json:"metrics"`
	UnprocessedMessages uint32            `json:"unprocessedMessages"`
}

func NewSystemgeConnectionClient(name string, commands []string, status int, metrics map[string]uint64, unprocessedMessages uint32) *SystemgeConnectionClient {
	return &SystemgeConnectionClient{
		Name:                name,
		Commands:            commands,
		Status:              status,
		Metrics:             metrics,
		UnprocessedMessages: unprocessedMessages,
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
