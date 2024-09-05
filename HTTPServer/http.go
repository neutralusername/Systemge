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
	name string

	status      int
	statusMutex sync.Mutex

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger

	config         *Config.HTTPServer
	httpServer     *http.Server
	blacklist      *Tools.AccessControlList
	whitelist      *Tools.AccessControlList
	requestCounter atomic.Uint64

	mux *CustomMux
}

func New(name string, config *Config.HTTPServer, handlers Handlers) *HTTPServer {
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
		blacklist: Tools.NewAccessControlList(config.TcpServerConfig.Blacklist),
		whitelist: Tools.NewAccessControlList(config.TcpServerConfig.Whitelist),
	}
	for pattern, handler := range handlers {
		server.AddRoute(pattern, handler)
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
	server.httpServer = nil

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
	server.mux.AddRoute(pattern, server.httpRequestWrapper(handlerFunc))
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
