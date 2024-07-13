package Oauth2

import (
	"Systemge/Http"
	"Systemge/Utilities"

	"golang.org/x/oauth2"
)

type Config struct {
	Port              int
	AuthPath          string
	AuthCallbackPath  string
	OAuth2Config      *oauth2.Config
	Logger            *Utilities.Logger
	TokenHandler      func(*Server, *oauth2.Token) (map[string]interface{}, error)
	SessionLifetimeMs int
	Randomizer        *Utilities.Randomizer
	Oauth2State       string
}

func (config *Config) New() *Server {
	server := &Server{
		sessionRequestChannel: make(chan *Oauth2SessionRequest),
		config:                config,

		sessions: make(map[string]*Session),
	}
	server.config.Oauth2State = server.config.Randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	server.httpServer = Http.New(config.Port, map[string]Http.RequestHandler{
		config.AuthPath:         oauth2Auth(server.config.OAuth2Config, server.config.Oauth2State),
		config.AuthCallbackPath: oauth2Callback(server.config.OAuth2Config, server.config.Oauth2State, server.config.Logger, server.sessionRequestChannel, "http://127.0.0.1:8080/", "http://google.at"),
	})
	return server
}
