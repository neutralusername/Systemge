package Dashboard

import "encoding/json"

type Metrics struct {
	Metrics map[string]uint64 `json:"metrics"`
	Name    string            `json:"name"`
}

func unmarshalMetrics(data string) (Metrics, error) {
	var m Metrics
	err := json.Unmarshal([]byte(data), &m)
	return m, err
}
