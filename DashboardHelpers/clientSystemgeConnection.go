package DashboardHelpers

import "encoding/json"

type SystemgeServerClient struct {
	Name                string            `json:"name"`
	Commands            []string          `json:"commands"`
	Status              int               `json:"status"`
	Metrics             map[string]uint64 `json:"metrics"`
	UnprocessedMessages uint32            `json:"unprocessedMessages"`
}

func NewSystemgeConnectionClient(name string, commands []string, status int, metrics map[string]uint64, unprocessedMessages uint32) *SystemgeServerClient {
	return &SystemgeServerClient{
		Name:                name,
		Commands:            commands,
		Status:              status,
		Metrics:             metrics,
		UnprocessedMessages: unprocessedMessages,
	}
}
func (client *SystemgeServerClient) Marshal() []byte {
	bytes, err := json.Marshal(client)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalSystemgeConnectionClient(bytes []byte) (*SystemgeServerClient, error) {
	var client SystemgeServerClient
	err := json.Unmarshal(bytes, &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
