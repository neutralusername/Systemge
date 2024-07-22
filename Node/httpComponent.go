package Node

import (
	"Systemge/Error"
	"Systemge/Http"
)

type httpComponent struct {
	application HTTPComponent
	server      *Http.Server
}

func (node *Node) startHTTPComponent() error {
	node.http = &httpComponent{
		application: node.application.(HTTPComponent),
	}
	node.http.server = Http.New(node.http.application.GetHTTPComponentConfig())
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
