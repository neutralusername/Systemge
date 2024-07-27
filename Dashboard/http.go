package Dashboard

import (
	"github.com/neutralusername/Systemge/Config"
	"net/http"
)

func (app *App) GetHTTPMessageHandlers() map[string]http.HandlerFunc {
	return app.httpMessageHandlers
}

func (app *App) GetHTTPComponentConfig() *Config.HTTP {
	return &Config.HTTP{
		Server: app.config.Server,
	}
}
