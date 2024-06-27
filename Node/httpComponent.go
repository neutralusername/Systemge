package Node

import "Systemge/Error"

func (node *Node) startHTTPComponent() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	httpServer := createHTTPServer(node.httpComponent.GetHTTPComponentConfig().Port, node.httpComponent.GetHTTPRequestHandlers())
	err := startHTTPServer(httpServer, node.httpComponent.GetHTTPComponentConfig().TlsCertPath, node.httpComponent.GetHTTPComponentConfig().TlsKeyPath)
	if err != nil {
		return Error.New("Error starting http server", err)
	}
	node.httpServer = httpServer
	return nil
}

func (node *Node) stopHTTPComponent() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	err := stopHTTPServer(node.httpServer)
	if err != nil {
		return Error.New("Error stopping http server", err)
	}
	node.httpServer = nil
	return nil
}
