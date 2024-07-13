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

func DiscordAuth(discordOAuth2Config *oauth2.Config, oauth2State string) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		url := discordOAuth2Config.AuthCodeURL(oauth2State)
		http.Redirect(responseWriter, httpRequest, url, http.StatusTemporaryRedirect)
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
