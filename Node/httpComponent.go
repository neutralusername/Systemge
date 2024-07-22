package Node

import (
	"Systemge/Error"
	"Systemge/HTTP"
)

type httpComponent struct {
	application HTTPComponent
	server      *HTTP.Server
}

func (node *Node) startHTTPComponent() error {
	node.http = &httpComponent{
		application: node.application.(HTTPComponent),
	}
	node.http.server = HTTP.New(node.http.application.GetHTTPComponentConfig(), node.http.application.GetHTTPMessageHandlers())
	err := node.http.server.Start()
	if err != nil {
		return Error.New("failed starting http server", err)
	}
	return nil
}

func (node *Node) stopHTTPComponent() error {
	http := node.http
	node.http = nil
	err := http.server.Stop()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	return nil
}
