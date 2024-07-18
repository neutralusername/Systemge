package Oauth2

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Node"
	"Systemge/Tools"
	"net/http"
	"sync"
)

type Server struct {
	node                  *Node.Node
	httpServer            *http.Server
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2
	randomizer            *Tools.Randomizer

	sessions   map[string]*session
	identities map[string]*session
	mutex      sync.Mutex

	stopChannel chan string
	isStarted   bool
}

func New(config Config.Oauth2) (*Server, error) {
	if config.TokenHandler == nil {
		return nil, Error.New("TokenHandler is required", nil)
	}
	if config.OAuth2Config == nil {
		return nil, Error.New("OAuth2Config is required", nil)
	}
	server := &Server{
		sessionRequestChannel: make(chan *oauth2SessionRequest),
		config:                &config,
		sessions:              make(map[string]*session),
		identities:            make(map[string]*session),
		randomizer:            Tools.NewRandomizer(config.RandomizerSeed),
	}
	return server, nil
}
