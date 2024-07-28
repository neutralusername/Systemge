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
	mux := newCustomMux()
	server := &Server{
		config: config,
		httpServer: &http.Server{
			Addr:    ":" + Helpers.IntToString(int(config.Server.Port)),
			Handler: mux,
		},
		blacklist: Tools.NewAccessControlList(config.Server.Blacklist),
		whitelist: Tools.NewAccessControlList(config.Server.Whitelist),
	}
	for pattern, handler := range handlers {
		mux.AddRoute(pattern, server.accessControllWrapper(handler))
	}
	return server
}

func (server *Server) AddRoute(pattern string, handlerFunc http.HandlerFunc) {
	server.httpServer.Handler.(*CustomMux).AddRoute(pattern, server.accessControllWrapper(handlerFunc))
}

func (server *Server) RemoveRoute(pattern string) {
	server.httpServer.Handler.(*CustomMux).RemoveRoute(pattern)
}

func (server *Server) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *Server) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *Server) Start() error {
	errorChannel := make(chan error)
	ended := false
	go func() {
		if server.config.Server.TlsCertPath != "" && server.config.Server.TlsKeyPath != "" {
			err := server.httpServer.ListenAndServeTLS(server.config.Server.TlsCertPath, server.config.Server.TlsKeyPath)
			if err != nil {
				if err != http.ErrServerClosed {
					panic(err)
				}
				if !ended {
					errorChannel <- err
				}
			}
		} else {
			err := server.httpServer.ListenAndServe()
			if err != nil {
				if err != http.ErrServerClosed {
					panic(err)
				}
				if !ended {
					errorChannel <- err
				}
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)
	select {
	case err := <-errorChannel:
		return Error.New("failed to start http server", err)
	default:
	}
	ended = true
	return nil
}

func (server *Server) Stop() error {
	err := server.httpServer.Close()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	return nil
}
