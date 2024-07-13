package Oauth2

import (
	"Systemge/Http"
	"Systemge/Utilities"

	"golang.org/x/oauth2"
)

type ServerConfig struct {
	Port                  int
	AuthPath              string
	AuthCallbackPath      string
	OAuth2Config          *oauth2.Config
	Logger                *Utilities.Logger
	SessionRequestHandler func(*Server, *oauth2.Token) (map[string]interface{}, error)
}

func (oauth2ServerConfig *ServerConfig) New() *Server {
	server := &Server{
		logger:                oauth2ServerConfig.Logger,
		sessionRequestChannel: make(chan *Oauth2SessionRequest),
		oauth2Config:          oauth2ServerConfig.OAuth2Config,
		tokenHandler:          oauth2ServerConfig.SessionRequestHandler,
		randomizer:            Utilities.NewRandomizer(Utilities.GetSystemTime()),

		sessions: make(map[string]*Session),
	}
	server.oauth2State = server.randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC)
	server.httpServer = Http.New(oauth2ServerConfig.Port, map[string]Http.RequestHandler{
		oauth2ServerConfig.AuthPath:         Oauth2(server.oauth2Config, server.oauth2State),
		oauth2ServerConfig.AuthCallbackPath: Oauth2Callback(server.oauth2Config, server.oauth2State, server.logger, server.sessionRequestChannel, "http://127.0.0.1:8080/", "http://google.at"),
	})
	return server
}
