package Oauth2Server

import (
	"net/http"
)

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
		if err := server.tokenHandler(server.config.OAuth2Config, token); err != nil {
			http.Redirect(responseWriter, httpRequest, server.config.CallbackFailureRedirectUrl, http.StatusMovedPermanently)
			return
		}
		http.Redirect(responseWriter, httpRequest, server.config.CallbackSuccessRedirectUrl, http.StatusMovedPermanently)
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
