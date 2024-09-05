package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

type Oauth2 struct {
	TcpServerConfig *TcpServer `json:"tcpServerConfig"` // *required*

	InfoLoggerPath    string `json:"infoLoggerPath"`    // *optional*
	WarningLoggerPath string `json:"warningLoggerPath"` // *optional*

	RandomizerSeed int64 `json:"randomizerSeed"` // default: 0

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

type RemoteCommand struct {
	TcpConnectionConfig *TcpConnection `json:"tcpConnectionConfig"` // *required*
	TcpClientConfig     *TcpClient     `json:"tcpClientConfig"`     // *required*
	MaxServerNameLength int            `json:"maxServerLength"`     // default: <=0 == unlimited (clients that attempt to send a name larger than this will be rejected)
}

func UnmarshalCommandClient(data string) *RemoteCommand {
	var commandClient RemoteCommand
	err := json.Unmarshal([]byte(data), &commandClient)
	if err != nil {
		return nil
	}
	return &commandClient
}

type CommandServer struct {
	SystemgeServerConfig *SystemgeServer `json:"systemgeServerConfig"` // *required*

	MaxClientNameLength int `json:"maxClientNameLength"` // default: 0 == unlimited (clients that attempt to send a name larger than this will be rejected)
}
