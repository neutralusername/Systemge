package httpServer

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

func (server *HTTPServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	server.sessionId = tools.GenerateRandomString(constants.SessionIdLength, tools.ALPHA_NUMERIC)

	if server.status != status.Stopped {
		return errors.New("failed to start http server")
	}
	server.status = status.Pending

	if server.config.TcpListenerConfig.Ip == "" {
		ip, err := net.LookupIP(server.config.TcpListenerConfig.Domain)
		if err != nil {
			return err
		}
		server.config.TcpListenerConfig.Ip = ip[0].String()
	}

	server.httpServer = &http.Server{
		MaxHeaderBytes:    int(server.config.MaxHeaderBytes),
		ReadHeaderTimeout: time.Duration(server.config.ReadHeaderTimeoutMs) * time.Millisecond,
		WriteTimeout:      time.Duration(server.config.WriteTimeoutMs) * time.Millisecond,

		Addr:    server.config.TcpListenerConfig.Ip + ":" + helpers.Uint16ToString(server.config.TcpListenerConfig.Port),
		Handler: server.mux,
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
		server.status = status.Stopped
		server.httpServer = nil
		return err
	default:
	}
	server.status = status.Started

	return nil
}
