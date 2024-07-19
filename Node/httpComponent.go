package Node

import (
	"Systemge/Error"
	"Systemge/Http"
	"net/http"
)

func (node *Node) startHTTPComponent() error {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	handlers := make(map[string]http.HandlerFunc)
	for pattern, handler := range node.GetHTTPComponent().GetHTTPRequestHandlers() {
		handlers[pattern] = Http.AccessControllWrapper(handler, node.httpBlacklist, node.httpWhitelist)
	}
	httpServer := Http.New(node.GetHTTPComponent().GetHTTPComponentConfig().Server.Port, handlers)
	err := Http.Start(httpServer, node.GetHTTPComponent().GetHTTPComponentConfig().Server.TlsCertPath, node.GetHTTPComponent().GetHTTPComponentConfig().Server.TlsKeyPath)
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
	err := Http.Stop(node.httpServer)
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	node.httpServer = nil
	node.httpStarted = false
	return nil
}
