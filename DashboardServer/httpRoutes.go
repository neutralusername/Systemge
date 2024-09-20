package DashboardServer

import (
	"net/http"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
)

func (server *Server) registerModuleHttpHandlers(connectedClient *connectedClient) {
	server.httpServer.AddRoute("/service/"+connectedClient.connection.GetName(), func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/service/"+connectedClient.connection.GetName(), http.FileServer(http.Dir(server.frontendPath))).ServeHTTP(w, r)
	})

	commands := connectedClient.page.GetCachedCommands()
	if commands == nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to get commands for connectedClient \"" + connectedClient.connection.GetName() + "\"")
		}
		return
	}

	server.httpServer.AddRoute("/service/"+connectedClient.connection.GetName()+"/command", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, err := r.Body.Read(body)
		if err != nil {
			http.Error(w, Error.New("Failed to read body", err).Error(), http.StatusInternalServerError)
			return
		}
		command, err := DashboardHelpers.UnmarshalCommand(string(body))
		if err != nil {
			http.Error(w, Error.New("Failed to unmarshal command", err).Error(), http.StatusBadRequest)
			return
		}
		if server.config.FrontendPassword != "" && command.Password != server.config.FrontendPassword {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}
		result, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_COMMAND, DashboardHelpers.NewCommand(command.Command, command.Args).Marshal())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})
	for command := range commands {
		server.httpServer.AddRoute("/"+connectedClient.connection.GetName()+"/command/"+command, func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			password := query.Get("password")
			if server.config.FrontendPassword != "" && password != server.config.FrontendPassword {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
				return
			}
			args := query["arg"]
			result, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_COMMAND, DashboardHelpers.NewCommand(command, args).Marshal())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(result))
		})
	}
}

func (server *Server) unregisterModuleHttpHandlers(connectedClient *connectedClient) {
	server.httpServer.RemoveRoute("/" + connectedClient.connection.GetName())

	commands := connectedClient.page.GetCachedCommands()
	if commands == nil {
		server.errorLogger.Log("Failed to get commands for connectedClient \"" + connectedClient.connection.GetName() + "\"")
		return
	}
	server.httpServer.RemoveRoute("/" + connectedClient.connection.GetName() + "/command")
	for command := range commands {
		server.httpServer.RemoveRoute("/" + connectedClient.connection.GetName() + "/command/" + command)
	}
}
