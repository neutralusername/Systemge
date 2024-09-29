package Oauth2Server

import (
	"net/http"

	"golang.org/x/oauth2"
)

type oauth2SessionRequest struct {
	token          *oauth2.Token
	sessionChannel chan<- *session
}

func (server *Server) oauth2AuthCallback() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		state := httpRequest.FormValue("state")
		if state != server.config.Oauth2State {
			http.Redirect(responseWriter, httpRequest, server.config.CallbackFailureRedirectUrl, http.StatusMovedPermanently)
			return
		}
		code := httpRequest.FormValue("code")
		token, err := server.config.OAuth2Config.Exchange(httpRequest.Context(), code)
		if err != nil {
			http.Redirect(responseWriter, httpRequest, server.config.CallbackFailureRedirectUrl, http.StatusMovedPermanently)
			return
		}
		sessionChannel := make(chan *session)
		oauth2SessionRequest := &oauth2SessionRequest{
			token:          token,
			sessionChannel: sessionChannel,
		}
		server.sessionRequestChannel <- oauth2SessionRequest
		session := <-sessionChannel
		if session == nil {
			http.Redirect(responseWriter, httpRequest, server.config.CallbackFailureRedirectUrl, http.StatusMovedPermanently)
			return
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
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
	}
}
