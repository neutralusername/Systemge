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

	HeapUpdateIntervalMs          uint64 `json:"heapUpdateIntervalMs"`          // default: 0 = disabled
	GoroutineUpdateIntervalMs     uint64 `json:"goroutineUpdateIntervalMs"`     // default: 0 = disabled
	ServiceStatusUpdateIntervalMs uint64 `json:"serviceStatusUpdateIntervalMs"` // default: 0 = disabled
	MetricsUpdateIntervalMs       uint64 `json:"metricsUpdateIntervalMs"`       // default: 0 = disabled
}

func UnmarshalDashboardServer(data string) *DashboardServer {
	var dashboard DashboardServer
	json.Unmarshal([]byte(data), &dashboard)
	return &dashboard
}
