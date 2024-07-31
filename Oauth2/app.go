package Oauth2

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
)

type App struct {
	node                  *Node.Node
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config.Oauth2

	sessions   map[string]*session
	identities map[string]*session
	mutex      sync.Mutex
}

func New(config *Config.Oauth2) (*Node.Node, error) {
	if config.TokenHandler == nil {
		return nil, Error.New("TokenHandler is required", nil)
	}
	if config.OAuth2Config == nil {
		return nil, Error.New("OAuth2Config is required", nil)
	}
	if config.NodeConfig == nil {
		return nil, Error.New("NodeConfig is required", nil)
	}
	app := &App{
		config:     config,
		sessions:   make(map[string]*session),
		identities: make(map[string]*session),
	}
	node := Node.New(&Config.NewNode{
		NodeConfig: config.NodeConfig,
		HttpConfig: &Config.HTTP{
			ServerConfig: config.ServerConfig,
		},
	}, app)
	return node, nil
}
