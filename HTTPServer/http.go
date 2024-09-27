package HTTPServer

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type Handlers map[string]http.HandlerFunc

type HTTPServer struct {
	name string

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
		name:      name,
		mux:       NewCustomMux(),
		config:    config,
		blacklist: blacklist,
		whitelist: whitelist,
	}
	for pattern, handler := range handlers {
		server.AddRoute(pattern, handler)
	}
	if config.ErrorLoggerPath != "" {
		file := Helpers.OpenFileAppend(config.ErrorLoggerPath)
		server.httpServer.ErrorLog = log.New(file, "[Error: \""+server.GetName()+"\"] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	return server
}

func (server *HTTPServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.StartingService,
		"Starting http server",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Start,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Stoped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"http server not stopped",
			Event.Context{
				Event.Circumstance: Event.Start,
			},
		))
		return errors.New("failed to start http server")
	}
	server.status = Status.Pending

	server.httpServer = &http.Server{
		MaxHeaderBytes:    int(server.config.MaxHeaderBytes),
		ReadHeaderTimeout: time.Duration(server.config.ReadHeaderTimeoutMs) * time.Millisecond,
		WriteTimeout:      time.Duration(server.config.WriteTimeoutMs) * time.Millisecond,

		Addr:    ":" + Helpers.IntToString(int(server.config.TcpServerConfig.Port)),
		Handler: server.mux,
	}

	errorChannel := make(chan error)
	ended := false
	go func() {
		if server.config.TcpServerConfig.TlsCertPath != "" && server.config.TcpServerConfig.TlsKeyPath != "" {
			err := server.httpServer.ListenAndServeTLS(server.config.TcpServerConfig.TlsCertPath, server.config.TcpServerConfig.TlsKeyPath)
			if err != nil {
				if !ended {
					errorChannel <- err
				} else if http.ErrServerClosed != err {
					if event := server.onEvent(Event.NewError(
						Event.UnexpectedError,
						err.Error(),
						Event.Panic,
						Event.Panic,
						Event.Cancel,
						Event.Context{
							Event.Circumstance: Event.Start,
						},
					)); !event.IsInfo() {
						panic(err)
					}
				}
			}
		} else {
			err := server.httpServer.ListenAndServe()
			if err != nil {
				if !ended {
					errorChannel <- err
				} else if http.ErrServerClosed != err {
					if event := server.onEvent(Event.NewError(
						Event.UnexpectedError,
						err.Error(),
						Event.Panic,
						Event.Panic,
						Event.Cancel,
						Event.Context{
							Event.Circumstance: Event.Start,
						},
					)); !event.IsInfo() {
						panic(err)
					}
				}
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)
	ended = true
	select {
	case err := <-errorChannel:
		server.status = Status.Stoped
		server.onEvent(Event.NewErrorNoOption(
			Event.UnexpectedError,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.Start,
			},
		))
		server.httpServer = nil
		server.status = Status.Stoped
		return err
	default:
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"http server started",
		Event.Context{
			Event.Circumstance: Event.Start,
		},
	))

	server.status = Status.Started
	return nil
}

func (server *HTTPServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.StoppingService,
		"Stopping http server",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Started {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"http server not started",
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
		return errors.New("http server not started")
	}
	server.status = Status.Pending

	err := server.httpServer.Close()
	if err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.UnexpectedError,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
	}
	server.httpServer = nil

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStopped,
		"http server stopped",
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	))
	server.status = Status.Stoped
	return nil
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

func (server *HTTPServer) GetName() string {
	return server.name
}

func (server *HTTPServer) GetStatus() int {
	return server.status
}

func (server *HTTPServer) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	if server.eventHandler == nil {
		return event
	}
	return server.eventHandler(event)
}
func (server *HTTPServer) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.TcpSystemgeListener,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.GetStatus()),
		Event.Function:      Event.GetCallerFuncName(2),
	}
}
