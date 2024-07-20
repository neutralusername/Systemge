package Dashboard

import (
	"Systemge/Config"
	"Systemge/Http"
	"net/http"
	"runtime"
)

func (app *App) GetHTTPRequestHandlers() map[string]http.HandlerFunc {
	_, file, _, _ := runtime.Caller(0)
	file = file[:len(file)-len("http.go")]
	return map[string]http.HandlerFunc{
		"/": Http.SendDirectory(file + "frontend"),
	}
}

func (app *App) GetHTTPComponentConfig() *Config.HTTP {
	return &Config.HTTP{
		Server: &Config.TcpServer{
			Port: app.config.HttpPort,
		},
	}
}
