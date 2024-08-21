package Dashboard

import "encoding/json"

type Metrics map[string]uint64

type metrics struct {
	Metrics Metrics `json:"metrics"`
	Name    string  `json:"name"`
}

func unmarshalMetrics(data string) (metrics, error) {
	var m metrics
	err := json.Unmarshal([]byte(data), &m)
	return m, err
}
