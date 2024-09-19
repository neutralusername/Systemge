package Metrics

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

type Metrics struct {
	KeyValuePairs map[string]uint64 `json:"keyValuePairs"`
	Time          time.Time         `json:"time"`
}

func New(keyValuePairs map[string]uint64) *Metrics {
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
	return Helpers.JsonMarshal(metrics)
}

func UnmarshalMetrics(data string) (*Metrics, error) {
	metrics := &Metrics{}
	err := json.Unmarshal([]byte(data), metrics)
	return metrics, err
}
