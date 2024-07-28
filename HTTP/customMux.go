package HTTP

import (
	"net/http"
	"sync"
)

type CustomMux struct {
	mux    *http.ServeMux
	routes map[string]http.Handler
	mutex  sync.Mutex
}

func newCustomMux() *CustomMux {
	return &CustomMux{
		mux:    http.NewServeMux(),
		routes: make(map[string]http.Handler),
	}
}

func (c *CustomMux) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.routes[pattern] = handlerFunc
	c.mux.Handle(pattern, handlerFunc)
}

func (c *CustomMux) RemoveRoute(pattern string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.routes[pattern]; ok {
		delete(c.routes, pattern)
		newMux := http.NewServeMux()
		for p, h := range c.routes {
			newMux.Handle(p, h)
		}
		c.mux = newMux
	}
}

func (c *CustomMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.mux.ServeHTTP(w, r)
}
