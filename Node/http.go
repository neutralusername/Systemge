package Node

import (
	"Systemge/Error"
	"net/http"
	"time"
)

func createHTTPServer(port string, handlers map[string]HTTPRequestHandler) *http.Server {
	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.HandleFunc(pattern, handler)
	}
	httpServer := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	return httpServer
}

func startHTTPServer(httpServer *http.Server, tlsCertPath, tlsKeyPath string) error {
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
		return Error.New("http server failed to start", err)
	default:
	}
	return nil
}

func stopHTTPServer(httpServer *http.Server) error {
	err := httpServer.Close()
	if err != nil {
		return Error.New("Error stopping http server", err)
	}
	return nil
}
