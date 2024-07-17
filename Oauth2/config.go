package Oauth2

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Utilities"
)

func New(config Config.Oauth2) (*Server, error) {
	if config.Randomizer == nil {
		config.Randomizer = Utilities.NewRandomizer(Utilities.GetSystemTime())
	}
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
	}
	return server, nil
}
