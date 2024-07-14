package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type oauth2SessionRequest struct {
	token            *oauth2.Token
	sessionIdChannel chan<- string
}

func (server *Server) oauth2Callback() Http.RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		server.config.Logger.Info("oauth2 callback for " + httpRequest.RemoteAddr)
		state := httpRequest.FormValue("state")
		if state != server.config.Oauth2State {
			server.config.Logger.Warning(Error.New("failed oauth2 state check", nil).Error())
			http.Redirect(responseWriter, httpRequest, server.config.FailureCallbackRedirect, http.StatusMovedPermanently)
			return
		}
		code := httpRequest.FormValue("code")
		token, err := server.config.OAuth2Config.Exchange(httpRequest.Context(), code)
		if err != nil {
			server.config.Logger.Warning(Error.New(fmt.Sprintf("failed exchanging code for token: %s", err.Error()), nil).Error())
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
			server.config.Logger.Warning(Error.New("failed creating session", nil).Error())
			http.Redirect(responseWriter, httpRequest, server.config.FailureCallbackRedirect, http.StatusMovedPermanently)
			return
		}
		server.config.Logger.Info("oauth2 callback success for " + httpRequest.RemoteAddr)
		http.Redirect(responseWriter, httpRequest, server.config.SucessCallbackRedirect+"?sessionId="+sessionId, http.StatusMovedPermanently)
	}
}

func (server *Server) oauth2Auth() Http.RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		server.config.Logger.Info("oauth2 request by " + httpRequest.RemoteAddr)
		url := server.config.OAuth2Config.AuthCodeURL(server.config.Oauth2State)
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
	}
}
