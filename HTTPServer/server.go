package HTTPServer

import (
	"Systemge/Application"
	"Systemge/Utilities"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	httpServer  *http.Server
	address     string
	tlsCertPath string
	tlsKeyPath  string
	name        string
	logger      *Utilities.Logger
	isStarted   bool
	mutex       sync.Mutex

	httpApplication Application.HTTPApplication
}

func New(port, name, tlsCertPath, tlsKeyPath string, logger *Utilities.Logger, httpApplication Application.HTTPApplication) *Server {
	server := &Server{
		httpServer:  nil,
		address:     port,
		tlsCertPath: tlsCertPath,
		tlsKeyPath:  tlsKeyPath,
		name:        name,
		logger:      logger,
		isStarted:   false,

		httpApplication: httpApplication,
	}
	return server
}

func (server *Server) SetKey(key string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Can not set key while server is running", nil)
	}
	server.tlsKeyPath = key
	return nil
}

func (server *Server) SetCert(cert string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.isStarted {
		return Utilities.NewError("Can not set cert while server is running", nil)
	}
	server.tlsCertPath = cert
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
	if server.httpApplication == nil {
		return Utilities.NewError("REST application not set", nil)
	}
	mux := http.NewServeMux()
	for pattern, handler := range server.httpApplication.GetHTTPRequestHandlers() {
		mux.HandleFunc(pattern, handler)
	}
	httpServer := &http.Server{
		Addr:    server.address,
		Handler: mux,
	}
	errorChannel := make(chan error)
	go func() {
		if server.tlsCertPath != "" && server.tlsKeyPath != "" {
			err := httpServer.ListenAndServeTLS(server.tlsCertPath, server.tlsKeyPath)
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

func (server *Server) GetAddress() string {
	return server.address
}

func (server *Server) GetName() string {
	return server.name
}
