package DashboardServer

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
)

func (server *Server) GetDefaultCommands() Commands.Handlers {
	dashboardHandlers := Commands.Handlers{
		"disconnectClient": func(args []string) (string, error) {
			if len(args) == 0 {
				return "", Error.New("No client name", nil)
			}
			if err := server.DisconnectClient(args[0]); err != nil {
				return "", err
			}
			return "success", nil
		},
	}
	if server.config.DashboardSystemgeCommands {
		systemgeDefaultCommands := server.systemgeServer.GetDefaultCommands()
		for command, handler := range systemgeDefaultCommands {
			server.dashboardCommandHandlers["systemgeServer_"+command] = handler
		}
	}
	if server.config.DashboardWebsocketCommands {
		httpDefaultCommands := server.httpServer.GetDefaultCommands()
		for command, handler := range httpDefaultCommands {
			server.dashboardCommandHandlers["httpServer_"+command] = handler
		}
	}
	if server.config.DashboardHttpCommands {
		webSocketDefaultCommands := server.websocketServer.GetDefaultCommands()
		for command, handler := range webSocketDefaultCommands {
			server.dashboardCommandHandlers["websocketServer_"+command] = handler
		}
	}
	server.dashboardClient = DashboardHelpers.NewDashboardClient(
		DashboardHelpers.DASHBOARD_CLIENT_NAME,
		server.dashboardCommandHandlers.GetKeyBoolMap(),
	)
	return dashboardHandlers
}
