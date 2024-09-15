package Metrics

import "time"

type Metrics struct {
	KeyValuePairs map[string]uint64
	Time          time.Time
}
