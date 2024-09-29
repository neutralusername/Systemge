package HTTPServer

import (
	"errors"
	"net/http"
	"time"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *HTTPServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"Starting http server",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"http server not stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
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
						Event.UnexpectedClosure,
						err.Error(),
						Event.Panic,
						Event.Panic,
						Event.Cancel,
						Event.Context{
							Event.Circumstance: Event.ServiceStart,
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
						Event.UnexpectedClosure,
						err.Error(),
						Event.Panic,
						Event.Panic,
						Event.Cancel,
						Event.Context{
							Event.Circumstance: Event.ServiceStart,
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
		server.status = Status.Stopped
		server.onEvent(Event.NewErrorNoOption(
			Event.ServiceStartFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		server.httpServer = nil
		server.status = Status.Stopped
		return err
	default:
	}
	server.status = Status.Started

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"http server started",
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	))

	return nil
}
