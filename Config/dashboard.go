package Config

import "encoding/json"

type DashboardServer struct {
	HTTPServerConfig      *HTTPServer      `json:"httpServerConfig"`      // *required*
	WebsocketServerConfig *WebsocketServer `json:"websocketServerConfig"` // *required*
	SystemgeServerConfig  *SystemgeServer  `json:"systemgeServerConfig"`  // *required*

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*

	MaxChartEntries uint32 `json:"maxChartEntries"` // default: 0 = disabled

	DashboardMetrics           bool `json:"dashboardMetrics"`           // default: false
	DashboardCommands          bool `json:"dashboardCommands"`          // default: false
	DashboardSystemgeCommands  bool `json:"dashboardSystemgeCommands"`  // default: false
	DashboardHttpCommands      bool `json:"dashboardHttpCommands"`      // default: false
	DashboardWebsocketCommands bool `json:"dashboardWebsocketCommands"` // default: false

	HeapUpdateIntervalMs      uint64 `json:"heapUpdateIntervalMs"`      // default: 0 = disabled
	GoroutineUpdateIntervalMs uint64 `json:"goroutineUpdateIntervalMs"` // default: 0 = disabled
	StatusUpdateIntervalMs    uint64 `json:"statusUpdateIntervalMs"`    // default: 0 = disabled
	MetricsUpdateIntervalMs   uint64 `json:"metricsUpdateIntervalMs"`   // default: 0 = disabled
}

func UnmarshalDashboardServer(data string) *DashboardServer {
	var dashboard DashboardServer
	err := json.Unmarshal([]byte(data), &dashboard)
	if err != nil {
		return nil
	}
	return &dashboard
}

type DashboardClient struct {
	MaxServerNameLength         int                    `json:"maxServerNameLength"`         // default: <=0 == no limit
	TcpSystemgeConnectionConfig *TcpSystemgeConnection `json:"tcpSystemgeConnectionConfig"` // *required*
	TcpClientConfig             *TcpClient             `json:"tcpClientConfig"`             // *required*
}

func UnmarshalDashboardClient(data string) *DashboardClient {
	var dashboard DashboardClient
	err := json.Unmarshal([]byte(data), &dashboard)
	if err != nil {
		return nil
	}
	return &dashboard
}
