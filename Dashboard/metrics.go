package Dashboard

import "encoding/json"

type metrics struct {
	Metrics map[string]uint64 `json:"metrics"`
	Name    string            `json:"name"`
}

func unmarshalMetrics(data string) (metrics, error) {
	var m metrics
	err := json.Unmarshal([]byte(data), &m)
	return m, err
}
