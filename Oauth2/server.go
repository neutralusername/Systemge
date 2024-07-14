package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type Server struct {
	httpServer            *http.Server
	sessionRequestChannel chan *oauth2SessionRequest
	config                *Config

	sessions   map[string]*session
	identities map[string]*session
	mutex      sync.Mutex

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
	server.config.Logger.Info(Error.New("started oauth2 server \""+server.config.Name+"\"", nil).Error())
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
	server.removeSessions()
	server.config.Logger.Info(Error.New("stopped oauth2 server \""+server.config.Name+"\"", nil).Error())
	return nil
}

func (server *Server) GetOauth2Config() *oauth2.Config {
	return server.config.OAuth2Config
}

func handleSessionRequests(server *Server) {
	for {
		select {
		case sessionRequest := <-server.sessionRequestChannel:
			server.config.Logger.Info(Error.New("handling session request with access token \""+sessionRequest.token.AccessToken+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
			handleSessionRequest(server, sessionRequest)
		case <-server.stopChannel:
			server.config.Logger.Info(Error.New("stopped handling session requests on oauth2 server \""+server.config.Name+"\"", nil).Error())
			return
		}
	}
}

func handleSessionRequest(server *Server, sessionRequest *oauth2SessionRequest) {
	identity, keyValuePairs, err := server.config.TokenHandler(server, sessionRequest.token)
	if err != nil {
		sessionRequest.sessionChannel <- nil
		server.config.Logger.Warning(Error.New("failed handling session request for access token \""+sessionRequest.token.AccessToken+"\" on oauth2 server \""+server.config.Name+"\"", err).Error())
		return
	}
	if identity == "" {
		sessionRequest.sessionChannel <- nil
		server.config.Logger.Warning(Error.New("no session identity for access token \""+sessionRequest.token.AccessToken+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
		return
	}
	sessionRequest.sessionChannel <- server.sessionIdentityFunc(identity, keyValuePairs)
}

// todo: find a proper name for this function
func (server *Server) sessionIdentityFunc(identity string, keyValuePairs map[string]interface{}) *session {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	session := server.identities[identity]
	if session == nil {
		session = server.createSession(identity, keyValuePairs)
		server.config.Logger.Info(Error.New("added session \""+session.sessionId+"\" with identity \""+session.identity+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
	} else {
		session.watchdog.Reset(time.Duration(server.config.SessionLifetimeMs) * time.Millisecond)
		session.expired = false
		server.config.Logger.Info(Error.New("refreshed session \""+session.sessionId+"\" with identity \""+session.identity+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
	}
	return session
}
