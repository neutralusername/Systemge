package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

// ServerConfig applies to both http and websocket besides the fact that websocket is hardcoded to port 18251
type Dashboard struct {
	ServerConfig *TcpServer `json:"serverConfig"` // *required*

	InfoLoggerPath    string  `json:"internalInfoLoggerPath"`    // *required*
	WarningLoggerPath string  `json:"internalWarningLoggerPath"` // *required*
	ErrorLoggerPath   string  `json:"errorLoggerPath"`           // *required*
	Mailer            *Mailer `json:"mailer"`                    // *required*

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
	ServerConfig *TcpServer `json:"serverConfig"` // *required*

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
