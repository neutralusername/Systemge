package Oauth2

import (
	"Systemge/Error"
	"Systemge/Node"
)

func (server *App) OnStart(node *Node.Node) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.node = node
	server.sessionRequestChannel = make(chan *oauth2SessionRequest)
	go handleSessionRequests(server)
	return nil
}

func (server *App) OnStop(node *Node.Node) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.node = nil
	close(server.sessionRequestChannel)
	server.removeAllSessions()
	return nil
}

func handleSessionRequests(server *App) {
	for sessionRequest := range server.sessionRequestChannel {
		if sessionRequest == nil {
			return
		}
		if infoLogger := server.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Handling session request with access token \""+sessionRequest.token.AccessToken+"\"", nil).Error())
		}
		handleSessionRequest(server, sessionRequest)
	}
}

func handleSessionRequest(server *App, sessionRequest *oauth2SessionRequest) {
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
