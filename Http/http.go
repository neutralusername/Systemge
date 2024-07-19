package Http

import (
	"Systemge/Error"
	"Systemge/Helpers"
	"net/http"
	"time"
)

func New(port uint16, handlers map[string]http.HandlerFunc) *http.Server {
	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.HandleFunc(pattern, handler)
	}
	httpServer := &http.Server{
		Addr:    ":" + Helpers.IntToString(int(port)),
		Handler: mux,
	}
	return httpServer
}

func Start(httpServer *http.Server, tlsCertPath, tlsKeyPath string) error {
	errorChannel := make(chan error)
	go func() {
		if tlsCertPath != "" && tlsKeyPath != "" {
			err := httpServer.ListenAndServeTLS(tlsCertPath, tlsKeyPath)
			if err != nil {
				if err != http.ErrServerClosed {
					panic(err)
				}
				errorChannel <- err
			}
		} else {
			err := httpServer.ListenAndServe()
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

func Stop(httpServer *http.Server) error {
	err := httpServer.Close()
	if err != nil {
		return Error.New("failed stopping http server", err)
	}
	return nil
}
