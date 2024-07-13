package Oauth2

import (
	"Systemge/Error"
	"Systemge/Http"
	"Systemge/Utilities"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type Oauth2SessionRequest struct {
	Token            *oauth2.Token
	SessionIdChannel chan string
}

func Oauth2Callback(oAuth2Config *oauth2.Config, oauth2State string, logger *Utilities.Logger, oauth2TokenChannel chan<- *Oauth2SessionRequest, successRedirectURL, failureRedirectURL string) Http.RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		state := httpRequest.FormValue("state")
		if state != oauth2State {
			logger.Warning(Error.New("oauth2 state mismatch", nil).Error())
			http.Redirect(responseWriter, httpRequest, failureRedirectURL, http.StatusMovedPermanently)
			return
		}
		code := httpRequest.FormValue("code")
		token, err := oAuth2Config.Exchange(httpRequest.Context(), code)
		if err != nil {
			logger.Warning(Error.New(fmt.Sprintf("failed exchanging code for token: %s", err.Error()), nil).Error())
			http.Redirect(responseWriter, httpRequest, failureRedirectURL, http.StatusMovedPermanently)
			return
		}
		Oauth2TokenRequest := &Oauth2SessionRequest{
			Token:            token,
			SessionIdChannel: make(chan string),
		}
		oauth2TokenChannel <- Oauth2TokenRequest
		sessionId := <-Oauth2TokenRequest.SessionIdChannel
		if sessionId == "" {
			http.Redirect(responseWriter, httpRequest, failureRedirectURL, http.StatusMovedPermanently)
			return
		}
		http.Redirect(responseWriter, httpRequest, successRedirectURL+"?sessionId="+sessionId, http.StatusMovedPermanently)
	}
}

func Oauth2(oAuth2Config *oauth2.Config, oauth2State string) Http.RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		url := oAuth2Config.AuthCodeURL(oauth2State)
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
	}
}
