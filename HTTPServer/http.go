package HTTPServer

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type Handlers map[string]http.HandlerFunc

type HTTPServer struct {
	status      int
	statusMutex sync.Mutex

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger

	config         *Config.HTTPServer
	httpServer     *http.Server
	blacklist      *Tools.AccessControlList
	whitelist      *Tools.AccessControlList
	handlers       Handlers
	requestCounter atomic.Uint64
}

func New(config *Config.HTTPServer, handlers Handlers) *HTTPServer {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpListenerConfig == nil {
		panic("config.TcpListenerConfig is nil")
	}
	mux := NewCustomMux()
	server := &HTTPServer{
		config:   config,
		handlers: handlers,
		httpServer: &http.Server{
			MaxHeaderBytes:    int(config.MaxHeaderBytes),
			ReadHeaderTimeout: time.Duration(config.ReadHeaderTimeoutMs) * time.Millisecond,
			WriteTimeout:      time.Duration(config.WriteTimeoutMs) * time.Millisecond,

			Addr:    ":" + Helpers.IntToString(int(config.TcpListenerConfig.Port)),
			Handler: mux,
		},
		blacklist: Tools.NewAccessControlList(config.TcpListenerConfig.Blacklist),
		whitelist: Tools.NewAccessControlList(config.TcpListenerConfig.Whitelist),
	}
	if config.ErrorLoggerPath != "" {
		file := Helpers.OpenFileAppend(config.ErrorLoggerPath)
		server.httpServer.ErrorLog = log.New(file, "[Error: \""+server.GetName()+"\"] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	if config.WarningLoggerPath != "" {
		server.warningLogger = Tools.NewLogger("[Warning: \""+server.GetName()+"\"] ", config.WarningLoggerPath)
	}
	if config.InfoLoggerPath != "" {
		server.infoLogger = Tools.NewLogger("[Info: \""+server.GetName()+"\"] ", config.InfoLoggerPath)
	}
	for pattern, handler := range handlers {
		handlers[pattern] = server.httpRequestWrapper(handler)
		mux.AddRoute(pattern, handlers[pattern])
	}
	return server
}

func (server *HTTPServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return Error.New("http server not stopped", nil)
	}
	server.status = Status.PENDING
	if server.infoLogger != nil {
		server.infoLogger.Log("starting http server")
	}

	errorChannel := make(chan error)
	ended := false
	go func() {
		if server.config.TcpListenerConfig.TlsCertPath != "" && server.config.TcpListenerConfig.TlsKeyPath != "" {
			err := server.httpServer.ListenAndServeTLS(server.config.TcpListenerConfig.TlsCertPath, server.config.TcpListenerConfig.TlsKeyPath)
			if err != nil {
				if !ended {
					errorChannel <- err
				} else if http.ErrServerClosed != err {
					panic(err)
				}
			}
		} else {
			err := server.httpServer.ListenAndServe()
			if err != nil {
				if !ended {
					errorChannel <- err
				} else if http.ErrServerClosed != err {
					panic(err)
				}
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)
	ended = true
	select {
	case err := <-errorChannel:
		server.status = Status.STOPPED
		return Error.New("failed to start http server", err)
	default:
	}

	if server.infoLogger != nil {
		server.infoLogger.Log("http server started")
	}
	server.status = Status.STARTED
	return nil
}

func (server *HTTPServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STARTED {
		return Error.New("http server not started", nil)
	}
	server.status = Status.PENDING
	if server.infoLogger != nil {
		server.infoLogger.Log("stopping http server")
	}

	err := server.httpServer.Close()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}

	if server.infoLogger != nil {
		server.infoLogger.Log("http server stopped")
	}
	server.status = Status.STOPPED
	return nil
}

func (server *HTTPServer) RetrieveHTTPRequestCounter() uint64 {
	return server.requestCounter.Swap(0)
}

func (server *HTTPServer) GetHTTPRequestCounter() uint64 {
	return server.requestCounter.Load()
}

func (server *HTTPServer) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	server.httpServer.Handler.(*CustomMux).AddRoute(pattern, handlerFunc)
}

func (server *HTTPServer) RemoveRoute(pattern string) {
	server.httpServer.Handler.(*CustomMux).RemoveRoute(pattern)
}

func (server *HTTPServer) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *HTTPServer) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *HTTPServer) GetName() string {
	return server.config.Name
}

func (server *HTTPServer) GetStatus() int {
	return server.status
}
