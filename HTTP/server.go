package HTTP

import (
	"Systemge/Utilities"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	httpServer *http.Server
	address    string
	tlsCert    string
	tlsKey     string
	name       string
	mux        *http.ServeMux
	logger     *Utilities.Logger
	isStarted  bool
	mutex      sync.Mutex
}

func New(port, name, tlsCert, tlsKey string, logger *Utilities.Logger) *Server {
	server := &Server{
		httpServer: nil,
		address:    port,
		tlsCert:    tlsCert,
		tlsKey:     tlsKey,
		name:       name,
		mux:        http.NewServeMux(),
		logger:     logger,
		isStarted:  false,
	}
	return server
}

func (server *Server) SetKey(key string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Can not set key while server is running", nil)
	}
	server.tlsKey = key
	return nil
}

func (server *Server) SetCert(cert string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Can not set cert while server is running", nil)
	}
	server.tlsCert = cert
	return nil
}

func (server *Server) IsStarted() bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.httpServer != nil
}

func (server *Server) Start() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Server already started", nil)
	}
	httpServer := &http.Server{
		Addr:    server.address,
		Handler: server.mux,
	}
	errorChannel := make(chan error)
	go func() {
		if server.tlsCert != "" && server.tlsKey != "" {
			err := httpServer.ListenAndServeTLS(server.tlsCert, server.tlsKey)
			if err != nil {
				if !server.isStarted {
					errorChannel <- err
				} else if err != http.ErrServerClosed {
					panic(err)
				}
			}
		} else {
			err := httpServer.ListenAndServe()
			if err != nil {
				if !server.isStarted {
					errorChannel <- err
				} else if err != http.ErrServerClosed {
					panic(err)
				}
			}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	select {
	case err := <-errorChannel:
		return Utilities.NewError("Server failed to start", err)
	default:
	}
	server.isStarted = true
	server.httpServer = httpServer
	return nil
}

func (server *Server) Stop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if !server.isStarted {
		return Utilities.NewError("Server not started", nil)
	}
	err := server.httpServer.Close()
	if err != nil {
		panic(err)
	}
	server.httpServer = nil
	server.isStarted = false
	return nil
}

// sets the mux for the server. the mux manages which handler corresponds to which pattern.
func (server *Server) SetMux(mux *http.ServeMux) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Can not set mux while server is running", nil)
	}
	server.mux = mux
	return nil
}

// resets the mux for the server. the mux manages which handler corresponds to which pattern.
func (server *Server) ResetMux() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Can not reset mux while server is running", nil)
	}
	server.mux = http.NewServeMux()
	return nil
}

// registers a handler for the given pattern
func (server *Server) RegisterPattern(pattern string, requestHandler RequestHandler) {
	server.mux.HandleFunc(pattern, requestHandler)
}

func (server *Server) GetAddress() string {
	return server.address
}

func (server *Server) GetName() string {
	return server.name
}
