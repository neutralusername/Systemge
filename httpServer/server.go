package httpServer

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

type HandlerFuncs map[string]http.HandlerFunc

type WrapperHandler func(http.ResponseWriter, *http.Request) error

type HTTPServer struct {
	config *configs.HTTPServer

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

func New(name string, config *configs.HTTPServer, wrapperHandler WrapperHandler, requestHandlers HandlerFuncs) (*HTTPServer, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpListenerConfig == nil {
		return nil, errors.New("config.TcpServerConfig is nil")
	}
	server := &HTTPServer{
		name:           name,
		mux:            NewCustomMux(config.DelayNs),
		config:         config,
		wrapperHandler: wrapperHandler,
		instanceId:     tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
	}
	for pattern, handler := range requestHandlers {
		server.AddRoute(pattern, handler)
	}
	if config.HttpErrorLogPath != "" {
		file := helpers.OpenFileAppend(config.HttpErrorLogPath)
		server.httpServer.ErrorLog = log.New(file, "[Error: \""+server.GetName()+"\"] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	return server, nil
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
