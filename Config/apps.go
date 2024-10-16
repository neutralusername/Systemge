package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

type Oauth2 struct {
	TcpServerConfig            *TcpServer                                                                  `json:"tcpServerConfig"`            // *required*
	InfoLoggerPath             string                                                                      `json:"infoLoggerPath"`             // *optional*
	WarningLoggerPath          string                                                                      `json:"warningLoggerPath"`          // *optional*
	RandomizerSeed             int64                                                                       `json:"randomizerSeed"`             // default: 0
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
	err := json.Unmarshal([]byte(data), &oauth2)
	if err != nil {
		return nil
	}
	return &oauth2
}
