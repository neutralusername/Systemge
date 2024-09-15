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

func ConvertMetrics(metrics map[string]map[string]uint64) map[string]map[string]*MetricsEntry {
	var metricsEntry map[string]map[string]*MetricsEntry = make(map[string]map[string]*MetricsEntry)
	for metricType, metricMap := range metrics {
		if metricsEntry[metricType] == nil {
			metricsEntry[metricType] = make(map[string]*MetricsEntry)
		}
		for metricName, metricValue := range metricMap {
			metricsEntry[metricType][metricName] = &MetricsEntry{
				Value: metricValue,
				Time:  time.Now(),
			}
		}
	}
	return metricsEntry
}

// merges metricsB into metricsA
func MergeMetrics(metricsA map[string]map[string]uint64, metricsB map[string]map[string]uint64) {
	for metricsType, metricsMap := range metricsB {
		metricsA[metricsType] = metricsMap
	}
}
