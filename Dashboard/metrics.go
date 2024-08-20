package Dashboard

import "encoding/json"

type Metrics map[string]uint64

func unmarshalMetrics(data string) (Metrics, error) {
	var m Metrics
	err := json.Unmarshal([]byte(data), &m)
	return m, err
}
