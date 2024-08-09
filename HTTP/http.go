package HTTP

import (
	"net/http"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	config     *Config.HTTP
	httpServer *http.Server
	blacklist  *Tools.AccessControlList
	whitelist  *Tools.AccessControlList
}

func New(config *Config.HTTP, handlers map[string]http.HandlerFunc) *Server {
	mux := NewCustomMux()
	server := &Server{
		config: config,
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
		mux.AddRoute(pattern, handler)
	}
	return server
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
