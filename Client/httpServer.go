package Client

import (
	"Systemge/Utilities"
)

func (client *Client) startApplicationHTTPServer() error {
	client.httpMutex.Lock()
	defer client.httpMutex.Unlock()
	httpServer := createHTTPServer(client.config.HTTPPort, client.httpApplication.GetHTTPRequestHandlers())
	err := startHTTPServer(httpServer, client.config.HTTPCertPath, client.config.HTTPKeyPath)
	if err != nil {
		return Utilities.NewError("Error starting http server", err)
	}
	client.httpServer = httpServer
	return nil
}

func (client *Client) stopApplicationHTTPServer() error {
	client.httpMutex.Lock()
	defer client.httpMutex.Unlock()
	err := stopHTTPServer(client.httpServer)
	if err != nil {
		return Utilities.NewError("Error stopping http server", err)
	}
	client.httpServer = nil
	return nil
}
