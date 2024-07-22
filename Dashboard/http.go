package Dashboard

import (
	"Systemge/Config"
	"Systemge/HTTP"
	"net/http"
	"runtime"
)

func (app *App) GetHTTPMessageHandlers() map[string]http.HandlerFunc {
	_, filePath, _, _ := runtime.Caller(0)
	return map[string]http.HandlerFunc{
		"/": HTTP.SendDirectory(filePath[:len(filePath)-len("http.go")] + "frontend"),
	}
}

func (app *App) GetHTTPComponentConfig() *Config.HTTP {
	return &Config.HTTP{
		Server: app.config.Server,
	}
}
