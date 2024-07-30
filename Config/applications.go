package Config

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

type Oauth2 struct {
	ServerConfig               *TcpServer                                                                  `json:"serverConfig"`               // *required*
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

type Spawner struct {
	PropagateSpawnedNodeChanges bool `json:"propagateSpawnedNodeChanges"` // default: false (if true, changes need to be received through the corresponding channel) (automated by dashboard)

	SpawnedNodeConfig *NewNode `json:"spawnedNodeConfig"` // *required* (the nodeId is attached to the node's name (e.g. "*nodeName*-*nodeId*"))
}

func UnmarshalSpawner(data string) *Spawner {
	var spawner Spawner
	json.Unmarshal([]byte(data), &spawner)
	return &spawner
}
