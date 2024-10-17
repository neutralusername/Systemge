package httpServer

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type CustomMux struct {
	routes  map[string]http.Handler
	mutex   sync.RWMutex
	delayNs int64
}

func NewCustomMux(delayNs int64) *CustomMux {
	return &CustomMux{
		routes:  make(map[string]http.Handler),
		delayNs: delayNs,
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

	time.Sleep(time.Duration(c.delayNs) * time.Nanosecond)

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
