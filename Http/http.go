package Http

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Helpers"
	"Systemge/Tools"
	"net/http"
	"time"
)

type Server struct {
	config     *Config.Http
	httpServer *http.Server
	blacklist  *Tools.AccessControlList
	whitelist  *Tools.AccessControlList
}

func (server *Server) GetWhitelist() *Tools.AccessControlList {
	return server.whitelist
}

func (server *Server) GetBlacklist() *Tools.AccessControlList {
	return server.blacklist
}

func (server *Server) GetHttpServer() *http.Server {
	return server.httpServer
}

func New(config *Config.Http) *Server {
	mux := http.NewServeMux()
	server := &Server{
		config: config,
		httpServer: &http.Server{
			Addr:    ":" + Helpers.IntToString(int(config.Server.Port)),
			Handler: mux,
		},
		blacklist: Tools.NewAccessControlList(config.Server.Blacklist),
		whitelist: Tools.NewAccessControlList(config.Server.Whitelist),
	}
	for pattern, handler := range config.Handlers {
		mux.HandleFunc(pattern, server.accessControllWrapper(handler))
	}
	return server
}

func (server *Server) Start() error {
	errorChannel := make(chan error)
	go func() {
		if server.config.Server.TlsCertPath != "" && server.config.Server.TlsKeyPath != "" {
			err := server.httpServer.ListenAndServeTLS(server.config.Server.TlsCertPath, server.config.Server.TlsKeyPath)
			if err != nil {
				if err != http.ErrServerClosed {
					panic(err)
				}
				errorChannel <- err
			}
		} else {
			err := server.httpServer.ListenAndServe()
			if err != nil {
				if err != http.ErrServerClosed {
					panic(err)
				}
				errorChannel <- err
			}
		}
	}()
	time.Sleep(100 * time.Millisecond)
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
