package Client

import (
	"Systemge/Application"
	"Systemge/HTTPServer"
	"Systemge/Resolution"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
)

func (client *Client) GetName() string {
	return client.name
}

func (client *Client) GetApplication() Application.Application {
	return client.application
}

func (client *Client) GetWebsocketServer() *WebsocketServer.Server {
	return client.websocketServer
}

func (client *Client) GetHTTPServer() *HTTPServer.Server {
	return client.httpServer
}

func (client *Client) GetResolverResolution() *Resolution.Resolution {
	return client.resolverResolution
}

func (client *Client) GetLogger() *Utilities.Logger {
	return client.logger
}

func (client *Client) GetHandleMessagesConcurrently() bool {
	client.handleMessagesConcurrentlyMutex.Lock()
	defer client.handleMessagesConcurrentlyMutex.Unlock()
	return client.handleMessagesConcurrently
}
