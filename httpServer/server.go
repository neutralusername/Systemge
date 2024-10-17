package httpServer

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/helpers"
	"github.com/neutralusername/Systemge/status"
	"github.com/neutralusername/Systemge/tools"
)

type HandlerFuncs map[string]http.HandlerFunc

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

	RequestCounter atomic.Uint64
}

func New(name string, config *Config.HTTPServer, wrapperHandler WrapperHandler, requestHandlers HandlerFuncs) *HTTPServer {
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
		instanceId:     tools.GenerateRandomString(Constants.InstanceIdLength, tools.ALPHA_NUMERIC),
	}
	for pattern, handler := range requestHandlers {
		server.AddRoute(pattern, handler)
	}
	if config.HttpErrorLogPath != "" {
		file := helpers.OpenFileAppend(config.HttpErrorLogPath)
		server.httpServer.ErrorLog = log.New(file, "[Error: \""+server.GetName()+"\"] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	return server
}

func (server *HTTPServer) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	server.mux.AddRoute(pattern, server.httpRequestWrapper(pattern, handlerFunc))
}
func (server *HTTPServer) httpRequestWrapper(pattern string, handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.statusMutex.RLock()
		defer server.statusMutex.RUnlock()
		if server.status != status.Started {
			Send403(w, r)
			return
		}

		server.RequestCounter.Add(1)
		r.Body = http.MaxBytesReader(w, r.Body, server.config.MaxBodyBytes)

		if server.wrapperHandler != nil {
			if err := server.wrapperHandler(w, r); err != nil {
				// do something with the error
				return
			}
		}

		handler(w, r)

	}
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