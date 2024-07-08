package Node

import "Systemge/Error"

func (node *Node) startHTTPComponent() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	httpServer := createHTTPServer(node.httpComponent.GetHTTPComponentConfig().Server.GetPort(), node.httpComponent.GetHTTPRequestHandlers())
	err := startHTTPServer(httpServer, node.httpComponent.GetHTTPComponentConfig().Server.GetTlsCertPath(), node.httpComponent.GetHTTPComponentConfig().Server.GetTlsKeyPath())
	if err != nil {
		return Error.New("failed starting http server", err)
	}
	node.httpServer = httpServer
	node.httpStarted = true
	return nil
}

func (node *Node) stopHTTPComponent() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	err := stopHTTPServer(node.httpServer)
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	node.httpServer = nil
	node.httpStarted = false
	return nil
}
