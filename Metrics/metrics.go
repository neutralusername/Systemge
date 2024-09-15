package Metrics

import "time"

type Metrics struct {
	KeyValuePairs map[string]uint64 `json:"keyValuePairs"`
	Time          time.Time         `json:"time"`
}

func Merge(typeMetricsA map[string]*Metrics, typeMetricsB map[string]*Metrics) {
	for metricsType, metricsMap := range typeMetricsB {
		typeMetricsA[metricsType] = metricsMap
	}
}
