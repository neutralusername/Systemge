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
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config

	sessions map[string]*session
	mutex    sync.Mutex

	stopChannel chan string
	isStarted   bool
}

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Error.New("oauth2 server \""+server.config.Name+"\" is already started", nil)
	}
	go handleSessionRequests(server)
	Http.Start(server.httpServer, "", "")
	server.isStarted = true
	server.config.Logger.Info("started oauth2 server \"" + server.config.Name + "\"")
	return nil
}

func (server *Server) Stop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if !server.isStarted {
		return Error.New("oauth2 server \""+server.config.Name+"\" is not started", nil)
	}
	server.httpServer.Close()
	server.isStarted = false
	close(server.stopChannel)
	for sessionId, session := range server.sessions {
		server.stop(session)
		delete(server.sessions, sessionId)
	}
	server.config.Logger.Info("stopped oauth2 server \"" + server.config.Name + "\"")
	return nil
}

func (server *Server) GetOauth2Config() *oauth2.Config {
	return server.config.OAuth2Config
}

func handleSessionRequests(server *Server) {
	for {
		select {
		case sessionRequest := <-server.sessionRequestChannel:
			server.config.Logger.Info("handling session request \"" + sessionRequest.token.AccessToken + "\"")
			handleSessionRequest(server, sessionRequest)
		case <-server.stopChannel:
			server.config.Logger.Info("stopped handling session requests")
			return
		}
	}
}

func handleSessionRequest(server *Server, sessionRequest *oauth2SessionRequest) {
	keyValuePairs, err := server.config.TokenHandler(server, sessionRequest.token)
	if err != nil {
		sessionRequest.sessionIdChannel <- ""
		server.config.Logger.Warning(Error.New("failed handling session request", err).Error())
		return
	}
	sessionRequest.sessionIdChannel <- server.addSession(newSession(keyValuePairs))
}
