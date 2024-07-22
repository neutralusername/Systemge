package Oauth2

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Node"
	"sync"
)

type App struct {
	node                  *Node.Node
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2

	sessions   map[string]*session
	identities map[string]*session
	mutex      sync.Mutex
}

func New(config *Config.Oauth2) (*App, error) {
	if config.TokenHandler == nil {
		return nil, Error.New("TokenHandler is required", nil)
	}
	if config.OAuth2Config == nil {
		return nil, Error.New("OAuth2Config is required", nil)
	}
	server := &App{
		config:     config,
		sessions:   make(map[string]*session),
		identities: make(map[string]*session),
	}
	return server, nil
}
