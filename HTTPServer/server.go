package HTTPServer

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tools"
)

type Handlers map[string]http.HandlerFunc

type HTTPServer struct {
	instanceId string
	sessionId  string
	name       string

	status      int
	statusMutex sync.RWMutex

	config        *Config.HTTPServer
	httpServer    *http.Server
	blacklist     *Tools.AccessControlList
	whitelist     *Tools.AccessControlList
	ipRateLimiter *Tools.IpRateLimiter

	eventHandler Event.Handler

	mux *CustomMux

	// metrics

	requestCounter atomic.Uint64
}

func New(name string, config *Config.HTTPServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, ipRateLimiter *Tools.IpRateLimiter, handlers Handlers, eventHandler *Event.Handler) *HTTPServer {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpServerConfig == nil {
		panic("config.TcpListenerConfig is nil")
	}
	server := &HTTPServer{
		name:          name,
		mux:           NewCustomMux(config.DelayNs),
		config:        config,
		instanceId:    Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		blacklist:     blacklist,
		whitelist:     whitelist,
		ipRateLimiter: ipRateLimiter,
	}
	for pattern, handler := range handlers {
		server.AddRoute(pattern, handler)
	}
	if config.HttpErrorPath != "" {
		file := Helpers.OpenFileAppend(config.HttpErrorPath)
		server.httpServer.ErrorLog = log.New(file, "[Error: \""+server.GetName()+"\"] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	return server
}

func (server *HTTPServer) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	server.mux.AddRoute(pattern, server.httpRequestWrapper(pattern, handlerFunc))
}

func (server *HTTPServer) RemoveRoute(pattern string) {
	server.mux.RemoveRoute(pattern)
}

func (server *HTTPServer) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *HTTPServer) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *HTTPServer) GetIpRateLimiter() *Tools.IpRateLimiter {
	return server.ipRateLimiter
}

func (server *HTTPServer) GetName() string {
	return server.name
}

func (server *HTTPServer) GetStatus() int {
	return server.status
}
