package HTTPServer

import (
	"net/http"
	"strings"
	"sync"
)

type CustomMux struct {
	routes map[string]http.Handler
	mutex  sync.RWMutex
}

func NewCustomMux() *CustomMux {
	return &CustomMux{
		routes: make(map[string]http.Handler),
	}
}

func (c *CustomMux) AddRoute(pattern string, handler http.Handler) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.routes[pattern] = handler
}

func (c *CustomMux) RemoveRoute(pattern string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.routes, pattern)
}

func (c *CustomMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if handler, ok := c.matchRoute(r.URL.Path); ok {
		handler.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func (c *CustomMux) matchRoute(path string) (http.Handler, bool) {
	if handler, ok := c.routes[path]; ok {
		return handler, true
	}

	for pattern, handler := range c.routes {
		if strings.HasPrefix(path, pattern) {
			return handler, true
		}
	}

	return nil, false
}
