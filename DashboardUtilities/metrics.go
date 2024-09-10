package DashboardUtilities

import "encoding/json"

type Metrics struct {
	Metrics map[string]uint64 `json:"metrics"`
	Name    string            `json:"name"`
}

func UnmarshalMetrics(data string) (Metrics, error) {
	var m Metrics
	err := json.Unmarshal([]byte(data), &m)
	return m, err
}
