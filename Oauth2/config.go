package Oauth2

import (
	"Systemge/Error"
	"Systemge/Utilities"

	"golang.org/x/oauth2"
)

type Config struct {
	Name                    string
	Port                    uint16
	AuthPath                string
	AuthCallbackPath        string
	OAuth2Config            *oauth2.Config
	SucessCallbackRedirect  string
	FailureCallbackRedirect string
	Logger                  *Utilities.Logger
	TokenHandler            func(*Server, *oauth2.Token) (string, map[string]interface{}, error)
	SessionLifetimeMs       uint64
	Randomizer              *Utilities.Randomizer
	Oauth2State             string
}

func (config Config) NewServer() (*Server, error) {
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

		sessions:   make(map[string]*session),
		identities: make(map[string]*session),
	}
	return server, nil
}
