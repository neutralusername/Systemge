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
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type Handlers map[string]http.HandlerFunc

type HTTPServer struct {
	instanceId string
	sessionId  string
	name       string

	status      int
	statusMutex sync.RWMutex

	config     *Config.HTTPServer
	httpServer *http.Server
	blacklist  *Tools.AccessControlList
	whitelist  *Tools.AccessControlList

	eventHandler Event.Handler

	mux *CustomMux

	// metrics

	requestCounter atomic.Uint64
}

func New(name string, config *Config.HTTPServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, handlers Handlers) *HTTPServer {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpServerConfig == nil {
		panic("config.TcpListenerConfig is nil")
	}
	server := &HTTPServer{
		name:       name,
		mux:        NewCustomMux(),
		config:     config,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		blacklist:  blacklist,
		whitelist:  whitelist,
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
	if event := server.onEvent(Event.NewInfo(
		Event.AddingRoute,
		"Adding route",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AddRoute,
			Event.Pattern:      pattern,
		},
	)); !event.IsInfo() {
		return
	}
	server.mux.AddRoute(pattern, server.httpRequestWrapper(pattern, handlerFunc))

	server.onEvent(Event.NewInfoNoOption(
		Event.AddedRoute,
		"Added route",
		Event.Context{
			Event.Circumstance: Event.AddRoute,
			Event.Pattern:      pattern,
		},
	))
}

func (server *HTTPServer) RemoveRoute(pattern string) {
	if event := server.onEvent(Event.NewInfo(
		Event.RemovingRoute,
		"Removing route",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.RemoveRoute,
			Event.Pattern:      pattern,
		},
	)); !event.IsInfo() {
		return
	}
	server.mux.RemoveRoute(pattern)

	server.onEvent(Event.NewInfoNoOption(
		Event.RemovedRoute,
		"Removed route",
		Event.Context{
			Event.Circumstance: Event.RemoveRoute,
			Event.Pattern:      pattern,
		},
	))
}

func (server *HTTPServer) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *HTTPServer) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *HTTPServer) GetName() string {
	return server.name
}

func (server *HTTPServer) GetStatus() int {
	return server.status
}

func (server *HTTPServer) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *HTTPServer) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.HttpServer,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
		Event.InstanceId:    server.instanceId,
		Event.SessionId:     server.sessionId,
	}
}
