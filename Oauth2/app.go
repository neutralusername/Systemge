package Oauth2

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Node"
	"Systemge/Tools"
	"net/http"
	"sync"
)

type App struct {
	node                  *Node.Node
	httpServer            *http.Server
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2

	sessions   map[string]*session
	identities map[string]*session
	mutex      sync.Mutex

	blacklist *Tools.AccessControlList
	whitelist *Tools.AccessControlList
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
		blacklist:  Tools.NewAccessControlList(),
		whitelist:  Tools.NewAccessControlList(),
	}
	for _, ip := range config.Blacklist {
		server.blacklist.Add(ip)
	}
	for _, ip := range config.Whitelist {
		server.whitelist.Add(ip)
	}
	return server, nil
}
