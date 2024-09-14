package DashboardHelpers

import "time"

type MetricsEntry struct {
	Value uint64    `json:"value"`
	Time  time.Time `json:"time"`
}
