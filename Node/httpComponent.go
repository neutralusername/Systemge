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
	if node.http == nil {
		return
	}
	handler = func(w http.ResponseWriter, r *http.Request) {
		node.http.requestCounter.Add(1)
		handler(w, r)
	}
	node.http.server.AddRoute(path, handler)
}

func (node *Node) RemoveHttpRoute(path string) {
	if node.http == nil {
		return
	}
	node.http.server.RemoveRoute(path)
}

func (node *Node) GetHTTPRequestCounter() uint64 {
	if node.http == nil {
		return 0
	}
	return node.http.requestCounter.Swap(0)
}

func (node *Node) startHTTPComponent() error {
	if node.newNodeConfig.HttpConfig == nil {
		return Error.New("http config is missing", nil)
	}
	node.http = &httpComponent{
		application: node.application.(HTTPComponent),
		config:      node.newNodeConfig.HttpConfig,
	}
	handlers := make(map[string]http.HandlerFunc)
	for path, handler := range node.http.application.GetHTTPMessageHandlers() {
		handlers[path] = func(w http.ResponseWriter, r *http.Request) {
			node.http.requestCounter.Add(1)
			handler(w, r)
		}
	}
	node.http.server = HTTP.New(node.http.config, handlers)
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
