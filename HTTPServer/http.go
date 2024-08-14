package HTTPServer

import (
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	config         *Config.HTTP
	httpServer     *http.Server
	blacklist      *Tools.AccessControlList
	whitelist      *Tools.AccessControlList
	handlers       map[string]http.HandlerFunc
	requestCounter atomic.Uint64
}

func (server *Server) httpRequestWrapper(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.requestCounter.Add(1)
		r.Body = http.MaxBytesReader(w, r.Body, server.config.MaxBodyBytes)

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			Send403(w, r)
			return
		}
		if server.GetBlacklist() != nil {
			if server.GetBlacklist().Contains(ip) {
				Send403(w, r)
				return
			}
		}
		if server.GetWhitelist() != nil && server.GetWhitelist().ElementCount() > 0 {
			if !server.GetWhitelist().Contains(ip) {
				Send403(w, r)
				return
			}
		}
		handler(w, r)
	}
}

func New(config *Config.HTTP, handlers map[string]http.HandlerFunc) *Server {
	mux := NewCustomMux()
	server := &Server{
		config:   config,
		handlers: handlers,
		httpServer: &http.Server{
			MaxHeaderBytes:    int(config.MaxHeaderBytes),
			ReadHeaderTimeout: time.Duration(config.ReadHeaderTimeoutMs) * time.Millisecond,
			WriteTimeout:      time.Duration(config.WriteTimeoutMs) * time.Millisecond,

			Addr:    ":" + Helpers.IntToString(int(config.ServerConfig.Port)),
			Handler: mux,
		},
		blacklist: Tools.NewAccessControlList(config.ServerConfig.Blacklist),
		whitelist: Tools.NewAccessControlList(config.ServerConfig.Whitelist),
	}
	for pattern, handler := range handlers {
		handlers[pattern] = server.httpRequestWrapper(handler)
		mux.AddRoute(pattern, handlers[pattern])
	}
	return server
}

func (server *Server) Start() error {
	errorChannel := make(chan error)
	ended := false
	go func() {
		if server.config.ServerConfig.TlsCertPath != "" && server.config.ServerConfig.TlsKeyPath != "" {
			err := server.httpServer.ListenAndServeTLS(server.config.ServerConfig.TlsCertPath, server.config.ServerConfig.TlsKeyPath)
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
		return Error.New("failed to start http server", err)
	default:
	}
	return nil
}

func (server *Server) Stop() error {
	err := server.httpServer.Close()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	return nil
}

func (server *Server) RetrieveHTTPRequestCounter() uint64 {
	return server.requestCounter.Swap(0)
}

func (server *Server) GetHTTPRequestCounter() uint64 {
	return server.requestCounter.Load()
}

func (server *Server) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	server.httpServer.Handler.(*CustomMux).AddRoute(pattern, handlerFunc)
}

func (server *Server) RemoveRoute(pattern string) {
	server.httpServer.Handler.(*CustomMux).RemoveRoute(pattern)
}

func (server *Server) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *Server) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}
