package Oauth2Server

import (
	"net/http"

	"github.com/neutralusername/Systemge/Error"

	"golang.org/x/oauth2"
)

type oauth2SessionRequest struct {
	token          *oauth2.Token
	sessionChannel chan<- *session
}

func (server *Server) oauth2AuthCallback() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Oauth2 callback for \""+httpRequest.RemoteAddr+"\" called", nil).Error())
		}
		state := httpRequest.FormValue("state")
		if state != server.config.Oauth2State {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed oauth2 state check for \""+state+"\" for client \""+httpRequest.RemoteAddr+"\"", nil).Error())
			}
			http.Redirect(responseWriter, httpRequest, server.config.CallbackFailureRedirectUrl, http.StatusMovedPermanently)
			return
		}
		code := httpRequest.FormValue("code")
		token, err := server.config.OAuth2Config.Exchange(httpRequest.Context(), code)
		if err != nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed exchanging code \""+code+"\" for token for client \""+httpRequest.RemoteAddr+"\"", err).Error())
			}
			http.Redirect(responseWriter, httpRequest, server.config.CallbackFailureRedirectUrl, http.StatusMovedPermanently)
			return
		}
		sessionChannel := make(chan *session)
		Oauth2SessionRequest := &oauth2SessionRequest{
			token:          token,
			sessionChannel: sessionChannel,
		}
		server.sessionRequestChannel <- Oauth2SessionRequest
		session := <-sessionChannel
		if session == nil {
			if warningLogger := server.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed creating session for access token \""+token.AccessToken+"\" for client \""+httpRequest.RemoteAddr+"\"", nil).Error())
			}
			http.Redirect(responseWriter, httpRequest, server.config.CallbackFailureRedirectUrl, http.StatusMovedPermanently)
			return
		}
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Created session for access token \""+token.AccessToken+"\" for client \""+httpRequest.RemoteAddr+"\" with sessionId \""+session.sessionId+"\" and identity \""+session.identity+"\"", nil).Error())
		}
		http.Redirect(responseWriter, httpRequest, server.config.CallbackSuccessRedirectUrl+"?sessionId="+session.sessionId, http.StatusMovedPermanently)
	}
}

func (server *Server) oauth2Auth() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		url := server.config.AuthRedirectUrl
		if url == "" {
			url = server.config.OAuth2Config.AuthCodeURL(server.config.Oauth2State)
		}
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log(Error.New("Oauth2 auth request by \""+httpRequest.RemoteAddr+"\"", nil).Error())
		}
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
	}
}
