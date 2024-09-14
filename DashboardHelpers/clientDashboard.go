package DashboardHelpers

import "encoding/json"

const (
	DASHBOARD_CLIENT_NAME               = "/"
	DASHBOARD_METRICSTYPE_SYSTEMGE      = "systemgeMetrics"
	DASHBOARD_METRICSTYPE_WEBSOCKET     = "websocketMetrics"
	DASHBOARD_METRICSTYPE_HTTP          = "httpMetrics"
	DASHBOARD_METRICSTYPE_RESOURCEUSAGE = "resourceUsageMetrics"
)

func NewDashboardClient(name string, commands map[string]bool) *DashboardClient {
	return &DashboardClient{
		Name:           name,
		Commands:       commands,
		ClientStatuses: map[string]int{},
	}
}

type DashboardClient struct {
	Name           string                                `json:"name"`
	Commands       map[string]bool                       `json:"commands"`
	ClientStatuses map[string]int                        `json:"clientStatuses"` //periodically automatically updated by the server
	Metrics        map[string]map[string][]*MetricsEntry `json:"metrics"`        //periodically automatically updated by the server
}

func (client *DashboardClient) Marshal() []byte {
	bytes, err := json.Marshal(client)
	if err != nil {
		panic(err)
	}
	return bytes
}
