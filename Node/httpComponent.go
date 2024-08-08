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
	server         *HTTP.Server
	requestCounter atomic.Uint64
}

func (node *Node) startHTTPComponent() (*httpComponent, error) {
	wrapperHandlers := make(map[string]http.HandlerFunc)
	http := &httpComponent{
		config: node.newNodeConfig.HttpConfig,
	}
	for path, handler := range node.application.(HTTPComponent).GetHTTPMessageHandlers() {
		wrapperHandlers[path] = http.httpRequestWrapper(handler)
	}
	http.server = HTTP.New(http.config, wrapperHandlers)
	err := http.server.Start()
	if err != nil {
		return nil, Error.New("failed starting http server", err)
	}
	return http, nil
}

func (node *Node) stopHTTPComponent() {
	http := node.http
	node.http = nil
	http.server.Stop()
}
