package Metrics

import "time"

type Metrics struct {
	KeyValuePairs map[string]uint64
	Time          time.Time
}

func Merge(typeMetricsA map[string]*Metrics, typeMetricsB map[string]*Metrics) {
	for metricsType, metricsMap := range typeMetricsB {
		typeMetricsA[metricsType] = metricsMap
	}
}
