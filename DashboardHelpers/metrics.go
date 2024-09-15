package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Metrics"
)

type DashboardMetrics map[string][]*Metrics.Metrics // metricType -> []*Metrics

func UnmarshalMetrics(data string) (DashboardMetrics, error) {
	var metrics DashboardMetrics
	err := json.Unmarshal([]byte(data), &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func NewDashboardMetrics(typeMetrics map[string]*Metrics.Metrics) DashboardMetrics {
	var dashboardMetrics = make(DashboardMetrics)
	for metricType, metrics := range typeMetrics {
		dashboardMetrics[metricType] = append(dashboardMetrics[metricType], metrics)
	}
	return dashboardMetrics
}

// merges metricsB into metricsA
func MergeMetrics(typeMetricsA map[string]*Metrics.Metrics, typeMetricsB map[string]*Metrics.Metrics) {
	for metricsType, metricsMap := range typeMetricsB {
		typeMetricsA[metricsType] = metricsMap
	}
}
