package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"Systemge/Node"
	"net/http"
)

func (server *Server) OnStart(node *Node.Node) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Error.New("oauth2 server \""+server.node.GetName()+"\" is already started", nil)
	}
	server.httpServer = Http.New(server.config.Server.Port, map[string]http.HandlerFunc{
		server.config.AuthPath:         Http.AccessControllWrapper(server.oauth2Auth(), &server.blacklist, &server.whitelist),
		server.config.AuthCallbackPath: Http.AccessControllWrapper(server.oauth2AuthCallback(), &server.blacklist, &server.whitelist),
	})
	err := Http.Start(server.httpServer, server.config.Server.TlsCertPath, server.config.Server.TlsKeyPath)
	if err != nil {
		return Error.New("failed to start oauth2 server \""+server.node.GetName()+"\"", err)
	}
	server.stopChannel = make(chan string)
	go handleSessionRequests(server)
	server.node = node
	server.isStarted = true
	return nil
}

func (server *Server) OnStop(node *Node.Node) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if !server.isStarted {
		return Error.New("oauth2 server \""+server.node.GetName()+"\" is not started", nil)
	}
	err := Http.Stop(server.httpServer)
	if err != nil {
		return Error.New("failed to stop oauth2 server \""+server.node.GetName()+"\"", err)
	}
	server.httpServer = nil
	server.node = nil
	server.isStarted = false
	close(server.stopChannel)
	server.removeAllSessions()
	return nil
}

func handleSessionRequests(server *Server) {
	for {
		select {
		case sessionRequest := <-server.sessionRequestChannel:
			if infoLogger := server.node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Handling session request with access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
			}
			handleSessionRequest(server, sessionRequest)
		case <-server.stopChannel:
			if infoLogger := server.node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Stopped handling session requests", nil).Error())
			}
			return
		}
	}
}

func handleSessionRequest(server *Server, sessionRequest *oauth2SessionRequest) {
	identity, keyValuePairs, err := server.config.TokenHandler(server.config.OAuth2Config, sessionRequest.token)
	if err != nil {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("Failed handling session request for access token \""+sessionRequest.token.AccessToken+"\"", err).Error())
		}
		return
	}
	if identity == "" {
		sessionRequest.sessionChannel <- nil
		if warningLogger := server.node.GetWarningLogger(); warningLogger != nil {
			warningLogger.Log(Error.New("No session identity for access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		return
	}
	sessionRequest.sessionChannel <- server.getSessionForIdentity(identity, keyValuePairs)
}
