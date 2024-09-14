package DashboardHelpers

import (
	"encoding/json"
	"time"
)

type MetricsEntry struct {
	Value uint64    `json:"value"`
	Time  time.Time `json:"time"`
}

func UnmarshalMetrics(data string) map[string]map[string]*MetricsEntry {
	var metrics map[string]map[string]*MetricsEntry
	err := json.Unmarshal([]byte(data), &metrics)
	if err != nil {
		return nil
	}
	return metrics
}
