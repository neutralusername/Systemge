package Node

import (
	"net"
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
		httpComponent.server.AddRoute(path, httpComponent.httpRequestWrapper(handler))
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

func (httpComponent *httpComponent) httpRequestWrapper(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpComponent.requestCounter.Add(1)
		r.Body = http.MaxBytesReader(w, r.Body, httpComponent.config.MaxBodyBytes)

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			HTTP.Send403(w, r)
			return
		}
		if httpComponent.server.GetBlacklist() != nil {
			if httpComponent.server.GetBlacklist().Contains(ip) {
				HTTP.Send403(w, r)
				return
			}
		}
		if httpComponent.server.GetWhitelist() != nil && httpComponent.server.GetWhitelist().ElementCount() > 0 {
			if !httpComponent.server.GetWhitelist().Contains(ip) {
				HTTP.Send403(w, r)
				return
			}
		}
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

func (node *Node) stopHTTPComponent() error {
	err := node.http.server.Stop()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	node.http = nil
	return nil
}
