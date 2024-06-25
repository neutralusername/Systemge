package Client

import (
	"Systemge/Utilities"
)

func (client *Client) StartApplicationHTTPServer() error {
	client.httpMutex.Lock()
	defer client.httpMutex.Unlock()
	httpServer := CreateHTTPServer(client.config.HTTPPort, client.config.HTTPCertPath, client.config.HTTPKeyPath, client.httpApplication.GetHTTPRequestHandlers())
	err := StartHTTPServer(httpServer, client.config.HTTPCertPath, client.config.HTTPKeyPath)
	if err != nil {
		return Utilities.NewError("Error starting http server", err)
	}
	client.httpServer = httpServer
	return nil
}

func (client *Client) StopApplicationHTTPServer() error {
	client.httpMutex.Lock()
	defer client.httpMutex.Unlock()
	err := StopHTTPServer(client.httpServer)
	if err != nil {
		return Utilities.NewError("Error stopping http server", err)
	}
	client.httpServer = nil
	return nil
}
