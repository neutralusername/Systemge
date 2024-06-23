package Client

import (
	"Systemge/Application"
	"Systemge/HTTPServer"
	"Systemge/Resolution"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
)

func (client *Client) SetApplication(application Application.Application) {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if !client.isStarted {
		client.application = application
	}
}

func (client *Client) SetWebsocketServer(websocketServer *WebsocketServer.Server) {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if !client.isStarted {
		client.websocketServer = websocketServer
	}
}

func (client *Client) SetHTTPServer(httpServer *HTTPServer.Server) {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if !client.isStarted {
		client.httpServer = httpServer
	}
}

func (client *Client) SetResolverResolution(resolverResolution *Resolution.Resolution) {
	client.resolverResolution = resolverResolution
}

func (client *Client) SetLogger(logger *Utilities.Logger) {
	client.logger = logger
}

func (client *Client) SetHandleMessagesConcurrently(handleMessagesConcurrently bool) {
	client.handleMessagesConcurrentlyMutex.Lock()
	defer client.handleMessagesConcurrentlyMutex.Unlock()
	client.handleMessagesConcurrently = handleMessagesConcurrently
}
