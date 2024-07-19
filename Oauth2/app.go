package Oauth2

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Node"
	"net/http"
	"sync"
)

type Server struct {
	node                  *Node.Node
	httpServer            *http.Server
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2

	sessions   map[string]*session
	identities map[string]*session
	mutex      sync.Mutex

	stopChannel chan string
	isStarted   bool

	blacklist map[string]bool
	whitelist map[string]bool
}

func New(config *Config.Oauth2) (*Server, error) {
	if config.TokenHandler == nil {
		return nil, Error.New("TokenHandler is required", nil)
	}
	if config.OAuth2Config == nil {
		return nil, Error.New("OAuth2Config is required", nil)
	}
	server := &Server{
		sessionRequestChannel: make(chan *oauth2SessionRequest),
		config:                config,
		sessions:              make(map[string]*session),
		identities:            make(map[string]*session),
		blacklist:             make(map[string]bool),
		whitelist:             make(map[string]bool),
	}
	for _, ip := range config.Blacklist {
		server.addToBlacklist(ip)
	}
	for _, ip := range config.Whitelist {
		server.addToWhitelist(ip)
	}
	return server, nil
}
