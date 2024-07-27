package Node

import (
	"net/http"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTP"
)

type httpComponent struct {
	application    HTTPComponent
	server         *HTTP.Server
	requestCounter atomic.Uint64
}

func (node *Node) GetHTTPRequestCounter() uint64 {
	if node.http == nil {
		return 0
	}
	return node.http.requestCounter.Swap(0)
}

func (node *Node) startHTTPComponent() error {
	node.http = &httpComponent{
		application: node.application.(HTTPComponent),
	}
	handlers := make(map[string]http.HandlerFunc)
	for path, handler := range node.http.application.GetHTTPMessageHandlers() {
		handlers[path] = func(w http.ResponseWriter, r *http.Request) {
			node.http.requestCounter.Add(1)
			handler(w, r)
		}
	}
	node.http.server = HTTP.New(node.http.application.GetHTTPComponentConfig(), handlers)
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
