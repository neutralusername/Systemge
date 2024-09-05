package Config

import "encoding/json"

type DashboardServer struct {
	HTTPServerConfig      *HTTPServer      `json:"httpServerConfig"`      // *required*
	WebsocketServerConfig *WebsocketServer `json:"websocketServerConfig"` // *required*
	SystemgeServerConfig  *SystemgeServer  `json:"systemgeServerConfig"`  // *required*

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *required*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *required*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *required*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *required*

	MaxChartEntries uint32 `json:"maxChartEntries"` // default: 0 = disabled

	HeapUpdateIntervalMs      uint64 `json:"heapUpdateIntervalMs"`          // default: 0 = disabled
	GoroutineUpdateIntervalMs uint64 `json:"goroutineUpdateIntervalMs"`     // default: 0 = disabled
	StatusUpdateIntervalMs    uint64 `json:"serviceStatusUpdateIntervalMs"` // default: 0 = disabled
	MetricsUpdateIntervalMs   uint64 `json:"metricsUpdateIntervalMs"`       // default: 0 = disabled
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
	MaxServerNameLength int            `json:"maxServerNameLength"` // default: <=0 = disabled
	ConnectionConfig    *TcpConnection `json:"connectionConfig"`    // *required*
	ClientConfig        *TcpClient     `json:"clientConfig"`        // *required*
}

func UnmarshalDashboardClient(data string) *DashboardClient {
	var dashboard DashboardClient
	err := json.Unmarshal([]byte(data), &dashboard)
	if err != nil {
		return nil
	}
	return &dashboard
}
