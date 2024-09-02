package Config

type SystemgeClient struct {
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ClientConfigs    []*TcpClient   `json:"clientConfigs"`    // *optional*
	ConnectionConfig *TcpConnection `json:"connectionConfig"` // *required*

	MaxServerNameLength int `json:"maxServerNameLength"` // default: 0 == unlimited (servers that attempt to send a name larger than this will be rejected)

	Reconnect                bool   `json:"reconnect"`               // default: false (if true, the client will attempt to reconnect if the connection is lost)
	ConnectionAttemptDelayMs uint64 `json:"reconnectAttemptDelayMs"` // default: 1000 (the delay between reconnection attempts in milliseconds)
	MaxConnectionAttempts    uint32 `json:"maxReconnectAttempts"`    // default: 0 == unlimited (the maximum number of reconnection attempts, after which the client will stop trying to reconnect)
}

type SystemgeServer struct {
	MailerConfig      *Mailer `json:"mailerConfig"`      // *optional*
	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *optional*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *optional*

	ListenerConfig   *TcpListener   `json:"listenerConfig"`   // *required*
	ConnectionConfig *TcpConnection `json:"connectionConfig"` // *required*
}
