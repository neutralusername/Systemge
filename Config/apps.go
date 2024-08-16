package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

type Dashboard struct {
	HTTPServerConfig      *HTTPServer       `json:"httpServerConfig"`      // *required*
	WebsocketServerConfig *WebsocketServer  `json:"websocketServerConfig"` // *required*
	SystemgeServerConfig  *SystemgeListener `json:"systemgeServerConfig"`  // *required*

	InfoLoggerPath    string  `json:"infoLoggerPath"`    // *required*
	WarningLoggerPath string  `json:"warningLoggerPath"` // *required*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`   // *required*
	MailerConfig      *Mailer `json:"mailerConfig"`      // *required*

	HeapUpdateIntervalMs          uint64 `json:"heapUpdateIntervalMs"`          // default: 0 = disabled
	GoroutineUpdateIntervalMs     uint64 `json:"goroutineUpdateIntervalMs"`     // default: 0 = disabled
	ServiceStatusUpdateIntervalMs uint64 `json:"serviceStatusUpdateIntervalMs"` // default: 0 = disabled
	MetricsUpdateIntervalMs       uint64 `json:"metricsUpdateIntervalMs"`       // default: 0 = disabled
}

func UnmarshalDashboard(data string) *Dashboard {
	var dashboard Dashboard
	json.Unmarshal([]byte(data), &dashboard)
	return &dashboard
}

type Oauth2 struct {
	ServerConfig *TcpListener `json:"serverConfig"` // *required*

	InfoLoggerPath    string `json:"internalInfoLoggerPath"`    // *required*
	WarningLoggerPath string `json:"internalWarningLoggerPath"` // *required*

	RandomizerSeed int64 `json:"randomizerSeed"` // *optional*

	AuthPath                   string                                                                      `json:"authPath"`                   // *required*
	AuthCallbackPath           string                                                                      `json:"authCallbackPath"`           // *required*
	OAuth2Config               *oauth2.Config                                                              `json:"oAuth2Config"`               // *required*
	AuthRedirectUrl            string                                                                      `json:"authRedirectUrl"`            // *optional*
	CallbackSuccessRedirectUrl string                                                                      `json:"callbackSuccessRedirectUrl"` // *required*
	CallbackFailureRedirectUrl string                                                                      `json:"callbackFailureRedirectUrl"` // *required*
	TokenHandler               func(*oauth2.Config, *oauth2.Token) (string, map[string]interface{}, error) `json:"-"`
	SessionLifetimeMs          uint64                                                                      `json:"sessionLifetimeMs"` // default: 0
	Oauth2State                string                                                                      `json:"oauth2State"`       // *required*
}

func UnmarshalOauth2(data string) *Oauth2 {
	var oauth2 Oauth2
	json.Unmarshal([]byte(data), &oauth2)
	return &oauth2
}
