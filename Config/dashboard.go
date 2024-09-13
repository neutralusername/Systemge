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

	MaxChartEntries             uint32 `json:"maxChartEntries"`             // default: 0 = disabled
	FrontendHeartbeatIntervalMs uint64 `json:"frontendHeartbeatIntervalMs"` // default: 0 = disabled

	DashboardMetrics           bool `json:"dashboardMetrics"`           // default: false
	DashboardSystemgeCommands  bool `json:"dashboardSystemgeCommands"`  // default: false
	DashboardHttpCommands      bool `json:"dashboardHttpCommands"`      // default: false
	DashboardWebsocketCommands bool `json:"dashboardWebsocketCommands"` // default: false

	UpdateIntervalMs uint64 `json:"updateIntervalMs"` // default: 1000
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
