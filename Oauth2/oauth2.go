package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type Server struct {
	httpServer            *http.Server
	sessionRequestChannel chan *Oauth2SessionRequest
	config                *Config

	sessions map[string]*Session
	mutex    sync.Mutex
}

func (server *Server) Start() {
	go handleSessionRequests(server)
	Http.Start(server.httpServer, "", "")
}

func (server *Server) GetOauth2Config() *oauth2.Config {
	return server.config.OAuth2Config
}

func handleSessionRequests(server *Server) {
	sessionRequest := <-server.sessionRequestChannel
	keyValuePairs, err := server.config.TokenHandler(server, sessionRequest.Token)
	if err != nil {
		sessionRequest.SessionIdChannel <- ""
		server.config.Logger.Warning(Error.New("failed handling session request", err).Error())
		return
	}
	sessionRequest.SessionIdChannel <- server.addSession(newSession(keyValuePairs))
}
