package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"net/http"

	"golang.org/x/oauth2"
)

type oauth2SessionRequest struct {
	token            *oauth2.Token
	sessionIdChannel chan<- string
}

func (server *Server) oauth2Callback() Http.RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		server.config.Logger.Info(Error.New("oauth2 callback for \""+httpRequest.RemoteAddr+"\"", nil).Error())
		state := httpRequest.FormValue("state")
		if state != server.config.Oauth2State {
			server.config.Logger.Warning(Error.New("failed oauth2 state check for \""+state+"\" for client \""+httpRequest.RemoteAddr+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
			http.Redirect(responseWriter, httpRequest, server.config.FailureCallbackRedirect, http.StatusMovedPermanently)
			return
		}
		code := httpRequest.FormValue("code")
		token, err := server.config.OAuth2Config.Exchange(httpRequest.Context(), code)
		if err != nil {
			server.config.Logger.Warning(Error.New("failed exchanging code \""+code+"\" for token for client \""+httpRequest.RemoteAddr+"\" on oauth2 server \""+server.config.Name+"\"", err).Error())
			http.Redirect(responseWriter, httpRequest, server.config.FailureCallbackRedirect, http.StatusMovedPermanently)
			return
		}
		sessionIdChannel := make(chan string)
		Oauth2TokenRequest := &oauth2SessionRequest{
			token:            token,
			sessionIdChannel: sessionIdChannel,
		}
		server.sessionRequestChannel <- Oauth2TokenRequest
		sessionId := <-sessionIdChannel
		if sessionId == "" {
			server.config.Logger.Warning(Error.New("failed creating session for access token \""+token.AccessToken+"\" for client \""+httpRequest.RemoteAddr+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
			http.Redirect(responseWriter, httpRequest, server.config.FailureCallbackRedirect, http.StatusMovedPermanently)
			return
		}
		server.config.Logger.Info(Error.New("oauth2 callback success for access token \""+token.AccessToken+"\" for client \""+httpRequest.RemoteAddr+"\" with sessionId \""+sessionId+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
		http.Redirect(responseWriter, httpRequest, server.config.SucessCallbackRedirect+"?sessionId="+sessionId, http.StatusMovedPermanently)
	}
}

func (server *Server) oauth2Auth() Http.RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		server.config.Logger.Info(Error.New("new oauth2 request by \""+httpRequest.RemoteAddr+"\" on oauth2 server \""+server.config.Name+"\"", nil).Error())
		url := server.config.OAuth2Config.AuthCodeURL(server.config.Oauth2State)
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
	}
}
