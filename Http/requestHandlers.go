package Http

import (
	"Systemge/Error"
	"Systemge/Tools"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
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

func WebsocketUpgrade(upgrader websocket.Upgrader, warningLogger *Tools.Logger, acceptConnection *bool, websocketConnChannel chan *websocket.Conn) RequestHandler {
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if !*acceptConnection {
			warningLogger.Log(Error.New("websocket connection not accepted", nil).Error())
			return
		}
		websocketConn, err := upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			warningLogger.Log(Error.New(fmt.Sprintf("failed upgrading connection to websocket: %s", err.Error()), nil).Error())
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
