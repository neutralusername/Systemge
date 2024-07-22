package Node

import (
	"Systemge/Error"
	"Systemge/Http"
)

func (node *Node) startHTTPComponent() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	httpServer := Http.New(node.GetHTTPComponent().GetHTTPComponentConfig())
	err := httpServer.Start()
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
	err := node.httpServer.Stop()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	node.httpServer = nil
	node.httpStarted = false
	return nil
}
