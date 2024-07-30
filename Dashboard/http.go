package Dashboard

import (
	"net/http"

	"github.com/neutralusername/Systemge/Config"
)

func (app *App) GetHTTPMessageHandlers() map[string]http.HandlerFunc {
	return nil
}

func (app *App) GetHTTPComponentConfig() *Config.HTTP {
	return &Config.HTTP{
		ServerConfig: app.config.ServerConfig,
	}
}
