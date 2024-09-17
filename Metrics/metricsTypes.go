package Metrics

type MetricsTypes map[string]*Metrics

func NewMetricsTypes() MetricsTypes {
	return make(MetricsTypes)
}

func (metricsTypes MetricsTypes) AddMetrics(metricsType string, metrics *Metrics) {
	metricsTypes[metricsType] = metrics
}

func (metricsTypes MetricsTypes) GetMetrics(metricsType string) *Metrics {
	return metricsTypes[metricsType]
}

func (metricsTypes MetricsTypes) RemoveMetrics(metricsType string) {
	delete(metricsTypes, metricsType)
}

func (metricsTypes MetricsTypes) GetMetricsTypes() []string {
	metricsTypesList := make([]string, 0, len(metricsTypes))
	for metricsType := range metricsTypes {
		metricsTypesList = append(metricsTypesList, metricsType)
	}
	return metricsTypesList
}

func (typeMetricsA MetricsTypes) Merge(typeMetricsB MetricsTypes) {
	for metricsType, metricsMap := range typeMetricsB {
		typeMetricsA[metricsType] = metricsMap
	}
}
