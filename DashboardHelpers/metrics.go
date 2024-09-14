package DashboardHelpers

import (
	"encoding/json"
	"time"
)

type MetricsEntry struct {
	Value uint64    `json:"value"`
	Time  time.Time `json:"time"`
}

func UnmarshalMetrics(data string) (map[string]map[string]*MetricsEntry, error) {
	var metrics map[string]map[string]*MetricsEntry
	err := json.Unmarshal([]byte(data), &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
