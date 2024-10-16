package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Metrics"
)

type DashboardMetrics map[string][]*Metrics.Metrics

func UnmarshalDashboardMetrics(data string) (DashboardMetrics, error) {
	var metrics DashboardMetrics
	err := json.Unmarshal([]byte(data), &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func NewDashboardMetrics(typeMetrics Metrics.MetricsTypes) DashboardMetrics {
	var dashboardMetrics = make(DashboardMetrics)
	for metricType, metrics := range typeMetrics {
		dashboardMetrics[metricType] = append(dashboardMetrics[metricType], metrics)
	}
	return dashboardMetrics
}
