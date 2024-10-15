package HTTPServer

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tools"
)

type Handlers map[string]http.HandlerFunc

type WrapperHandler func(http.ResponseWriter, *http.Request) error

type HTTPServer struct {
	config *Config.HTTPServer

	name       string
	instanceId string
	sessionId  string

	status      int
	statusMutex sync.RWMutex

	httpServer     *http.Server
	wrapperHandler WrapperHandler
	mux            *CustomMux

	// metrics

	requestCounter atomic.Uint64
}

func New(name string, config *Config.HTTPServer, wrapperHandler WrapperHandler, handlers Handlers) *HTTPServer {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpServerConfig == nil {
		panic("config.TcpListenerConfig is nil")
	}
	server := &HTTPServer{
		name:           name,
		mux:            NewCustomMux(config.DelayNs),
		config:         config,
		wrapperHandler: wrapperHandler,
		instanceId:     Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}
	for pattern, handler := range handlers {
		server.AddRoute(pattern, handler)
	}
	if config.HttpErrorLogPath != "" {
		file := Helpers.OpenFileAppend(config.HttpErrorLogPath)
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

func (server *HTTPServer) GetName() string {
	return server.name
}

func (server *HTTPServer) GetStatus() int {
	return server.status
}
