package Metrics

import "time"

type Metrics struct {
	KeyValue map[string]uint64
	Time     time.Time
}
