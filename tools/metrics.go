package tools

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/helpers"
)

type Metrics struct {
	KeyValuePairs map[string]uint64 `json:"keyValuePairs"`
	Time          time.Time         `json:"time"`
}

func NewMetrics(keyValuePairs map[string]uint64) *Metrics {
	return &Metrics{
		KeyValuePairs: keyValuePairs,
		Time:          time.Now(),
	}
}

func (metrics *Metrics) Add(key string, value uint64) {
	metrics.KeyValuePairs[key] = value
}

func (metrics *Metrics) Get(key string) uint64 {
	return metrics.KeyValuePairs[key]
}

func (metrics *Metrics) GetTime() time.Time {
	return metrics.Time
}

func (metrics *Metrics) Marshal() string {
	return helpers.JsonMarshal(metrics)
}

func UnmarshalMetrics(data string) (*Metrics, error) {
	metrics := &Metrics{}
	err := json.Unmarshal([]byte(data), metrics)
	return metrics, err
}

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

func (typeMetrics MetricsTypes) Marshal() string {
	return helpers.JsonMarshal(typeMetrics)
}

func UnmarshalMetricsTypes(data string) (MetricsTypes, error) {
	metricsTypes := MetricsTypes{}
	err := json.Unmarshal([]byte(data), &metricsTypes)
	return metricsTypes, err
}
