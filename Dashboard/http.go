package Dashboard

import (
	"Systemge/Config"
	"Systemge/Http"
	"net/http"
	"runtime"
)

func (app *App) GetHTTPRequestHandlers() map[string]http.HandlerFunc {
	_, filePath, _, _ := runtime.Caller(0)
	filePath = filePath[:len(filePath)-len("http.go")]
	return map[string]http.HandlerFunc{
		app.config.Pattern: Http.SendDirectory(filePath + "frontend"),
	}
}

func (app *App) GetHTTPComponentConfig() *Config.HTTP {
	return &Config.HTTP{
		Server: &Config.TcpServer{
			Port:        app.config.Server.Port,
			TlsCertPath: app.config.Server.TlsCertPath,
			TlsKeyPath:  app.config.Server.TlsKeyPath,
		},
	}
}
