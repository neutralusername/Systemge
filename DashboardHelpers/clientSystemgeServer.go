package DashboardHelpers

import "encoding/json"

type SystemgeServerlient struct {
	Name     string           `json:"name"`
	Commands map[string]bool  `json:"commands"`
	Status   int              `json:"status"`  //periodically automatically updated by the server
	Metrics  DashboardMetrics `json:"metrics"` //periodically automatically updated by the server
}

func NewSystemgeServerClient(name string, commands map[string]bool, status int, metrics DashboardMetrics) *SystemgeServerlient {
	return &SystemgeServerlient{
		Name:     name,
		Commands: commands,
		Status:   status,
		Metrics:  metrics,
	}
}
func (client *SystemgeServerlient) Marshal() []byte {
	bytes, err := json.Marshal(client)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalSystemgeServerlient(bytes []byte) (*SystemgeServerlient, error) {
	var client SystemgeServerlient
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
