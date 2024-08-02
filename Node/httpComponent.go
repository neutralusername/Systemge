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

func (node *Node) AddHttpRoute(path string, handler http.HandlerFunc) {
	if httpComponent := node.http; httpComponent != nil {
		httpComponent.server.AddRoute(path, httpComponent.counterWrapper(handler))
	}
}

func (node *Node) RemoveHttpRoute(path string) {
	if httpComponent := node.http; httpComponent != nil {
		httpComponent.server.RemoveRoute(path)
	}
}

func (node *Node) GetHTTPRequestCounter() uint64 {
	if httpComponent := node.http; httpComponent != nil {
		return httpComponent.requestCounter.Swap(0)
	}
	return 0
}

func (httpComponent *httpComponent) counterWrapper(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpComponent.requestCounter.Add(1)
		handler(w, r)
	}
}

func (node *Node) startHTTPComponent() error {
	if node.newNodeConfig.HttpConfig == nil {
		return Error.New("http config is missing", nil)
	}
	node.http = &httpComponent{
		application: node.application.(HTTPComponent),
		config:      node.newNodeConfig.HttpConfig,
	}
	counterWrappedHandlers := make(map[string]http.HandlerFunc)
	for path, handler := range node.http.application.GetHTTPMessageHandlers() {
		counterWrappedHandlers[path] = node.http.counterWrapper(handler)
	}
	node.http.server = HTTP.New(node.http.config, counterWrappedHandlers)
	err := node.http.server.Start()
	if err != nil {
		return Error.New("failed starting http server", err)
	}
	return nil
}

func (node *Node) stopHTTPComponent() error {
	err := node.http.server.Stop()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	node.http = nil
	return nil
}
