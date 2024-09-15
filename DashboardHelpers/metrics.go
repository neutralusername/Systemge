package DashboardHelpers

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/Metrics"
)

type DashboardMetrics map[string][]*Metrics.Metrics // metricType -> []*Metrics

func UnmarshalMetrics(data string) (map[string][]*Metrics.Metrics, error) {
	var metrics map[string][]*Metrics.Metrics
	err := json.Unmarshal([]byte(data), &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func ConvertToDashboardMetrics(metrics map[string]map[string]uint64) map[string]map[time.Time]map[string]uint64 {
	var metricsEntry map[string]map[time.Time]map[string]uint64
	for metricType, metricMap := range metrics {
		if metricsEntry[metricType] == nil {
			metricsEntry[metricType] = make(map[time.Time]map[string]uint64)
		}
		metricsEntry[metricType][time.Now()] = metricMap
	}
	return metricsEntry
}

// merges metricsB into metricsA
func MergeMetrics(metricsA map[string]map[string]uint64, metricsB map[string]map[string]uint64) {
	for metricsType, metricsMap := range metricsB {
		metricsA[metricsType] = metricsMap
	}
}
