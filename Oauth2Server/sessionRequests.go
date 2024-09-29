package Oauth2Server

import "context"

func (server *Server) handleSessionRequests() {
	for sessionRequest := range server.sessionRequestChannel {
		server.handleSessionRequest(sessionRequest)
	}
}

func (server *Server) handleSessionRequest(sessionRequest *oauth2SessionRequest) {
	client := server.config.OAuth2Config.Client(context.Background(), sessionRequest.token)
	identity, err := server.config.TokenHandler(server.config.OAuth2Config, client)
	if err != nil {
		sessionRequest.sessionChannel <- nil
		return
	}
	if identity == "" {
		sessionRequest.sessionChannel <- nil
		return
	}
	sessionRequest.sessionChannel <- server.getSessionForIdentity(identity)
}
