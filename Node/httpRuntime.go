package Node

import "net/http"

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
