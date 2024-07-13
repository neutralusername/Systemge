package Http

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

func SendDirectory(path string) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		http.FileServer(http.Dir(path)).ServeHTTP(responseWriter, httpRequest)
	}
}

func RedirectTo(toURL string) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		http.Redirect(responseWriter, httpRequest, toURL, http.StatusMovedPermanently)
	}
}

func Oauth2(oAuth2Config *oauth2.Config, oauth2State string) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		url := oAuth2Config.AuthCodeURL(oauth2State)
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
	}
}

type Oauth2SessionRequest struct {
	Token            *oauth2.Token
	SessionIdChannel chan string
}

func Oauth2Callback(oAuth2Config *oauth2.Config, oauth2State string, logger *Utilities.Logger, oauth2TokenChannel chan<- *Oauth2SessionRequest, successRedirectURL, failureRedirectURL string) RequestHandler {
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

func WebsocketUpgrade(upgrader websocket.Upgrader, logger *Utilities.Logger, acceptConnection *bool, websocketConnChannel chan *websocket.Conn) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if !*acceptConnection {
			logger.Warning(Error.New("websocket connection not accepted", nil).Error())
			return
		}
		websocketConn, err := upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			logger.Warning(Error.New(fmt.Sprintf("failed upgrading connection to websocket: %s", err.Error()), nil).Error())
			return
		}
		websocketConnChannel <- websocketConn
	}
}

func SendHTTPResponseFull(statusCode int, headerKeyValuePairs map[string]string, body string) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		for key, value := range headerKeyValuePairs {
			responseWriter.Header().Set(key, value)
		}
		responseWriter.WriteHeader(statusCode)
		responseWriter.Write([]byte(body))
	}
}

func SendHTTPResponseCodeAndBody(statusCode int, body string) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(statusCode)
		responseWriter.Write([]byte(body))
	}
}

func SendHTTPResponseCode(statusCode int) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		responseWriter.WriteHeader(statusCode)
	}
}
