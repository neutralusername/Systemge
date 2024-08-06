package Node

import "net/http"

// AddHttpRoute adds a route to the http server.
func (node *Node) AddHttpRoute(path string, handler http.HandlerFunc) {
	if httpComponent := node.http; httpComponent != nil {
		httpComponent.server.AddRoute(path, httpComponent.httpRequestWrapper(handler))
	}
}

// RemoveHttpRoute removes a route from the http server.
func (node *Node) RemoveHttpRoute(path string) {
	if httpComponent := node.http; httpComponent != nil {
		httpComponent.server.RemoveRoute(path)
	}
}
