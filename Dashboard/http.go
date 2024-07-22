package Dashboard

import (
	"Systemge/Config"
	"Systemge/Http"
	"net/http"
	"runtime"
)

func (app *App) GetHTTPComponentConfig() *Config.Http {
	_, filePath, _, _ := runtime.Caller(0)
	return &Config.Http{
		Server: app.config.Server,
		Handlers: map[string]http.HandlerFunc{
			"/": Http.SendDirectory(filePath[:len(filePath)-len("http.go")] + "frontend"),
		},
	}
}
