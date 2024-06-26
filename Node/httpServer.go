package Node

import "Systemge/Error"

func (node *Node) startApplicationHTTPServer() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	httpServer := createHTTPServer(node.config.HTTPPort, node.httpComponent.GetHTTPRequestHandlers())
	err := startHTTPServer(httpServer, node.config.HTTPCertPath, node.config.HTTPKeyPath)
	if err != nil {
		return Error.New("Error starting http server", err)
	}
	node.httpServer = httpServer
	return nil
}

func (node *Node) stopApplicationHTTPServer() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	err := stopHTTPServer(node.httpServer)
	if err != nil {
		return Error.New("Error stopping http server", err)
	}
	node.httpServer = nil
	return nil
}
