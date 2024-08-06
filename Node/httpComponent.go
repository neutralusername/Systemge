package Node

import (
	"net/http"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTP"
)

type httpComponent struct {
	config         *Config.HTTP
	application    HTTPComponent
	server         *HTTP.Server
	requestCounter atomic.Uint64
}

func (node *Node) startHTTPComponent() error {
	node.http = &httpComponent{
		application: node.application.(HTTPComponent),
		config:      node.newNodeConfig.HttpConfig,
	}
	wrapperHandlers := make(map[string]http.HandlerFunc)
	for path, handler := range node.http.application.GetHTTPMessageHandlers() {
		wrapperHandlers[path] = node.http.httpRequestWrapper(handler)
	}
	node.http.server = HTTP.New(node.http.config, wrapperHandlers)
	err := node.http.server.Start()
	if err != nil {
		return Error.New("failed starting http server", err)
	}
	return nil
}

func (node *Node) stopHTTPComponent() {
	http := node.http
	node.http = nil
	http.server.Stop()
}
