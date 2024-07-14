package Oauth2

import (
	"Systemge/Http"
	"Systemge/Utilities"

	"golang.org/x/oauth2"
)

type Config struct {
	Port                    int
	AuthPath                string
	AuthCallbackPath        string
	OAuth2Config            *oauth2.Config
	SucessCallbackRedirect  string
	FailureCallbackRedirect string
	Logger                  *Utilities.Logger
	TokenHandler            func(*Server, *oauth2.Token) (map[string]interface{}, error)
	SessionLifetimeMs       int
	Randomizer              *Utilities.Randomizer
	Oauth2State             string
}

func (config *Config) New() *Server {
	server := &Server{
		sessionRequestChannel: make(chan *oauth2SessionRequest),
		config:                config,

		sessions: make(map[string]*session),
	}
	server.config.Oauth2State = server.config.Randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	server.httpServer = Http.New(config.Port, map[string]Http.RequestHandler{
		config.AuthPath:         server.oauth2Auth(),
		config.AuthCallbackPath: server.oauth2Callback(),
	})
	return server
}
