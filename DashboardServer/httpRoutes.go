package DashboardServer

import (
	"net/http"
	"strings"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
)

func (server *Server) registerModuleHttpHandlers(connectedClient *connectedClient) {
	server.httpServer.AddRoute("/"+connectedClient.connection.GetName(), func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+connectedClient.connection.GetName(), http.FileServer(http.Dir(server.frontendPath))).ServeHTTP(w, r)
	})

	commands := DashboardHelpers.GetCommands(connectedClient.client)
	if commands == nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to get commands for connectedClient \"" + connectedClient.connection.GetName() + "\"")
		}
		return
	}

	for command := range commands {
		server.httpServer.AddRoute("/"+connectedClient.connection.GetName()+"/command/"+command, func(w http.ResponseWriter, r *http.Request) {
			body := make([]byte, r.ContentLength)
			_, err := r.Body.Read(body)
			if err != nil {
				http.Error(w, Error.New("Failed to read body", err).Error(), http.StatusInternalServerError)
				return
			}
			args := strings.Split(string(body), " ")
			if len(args) == 0 {
				http.Error(w, "No command", http.StatusBadRequest)
				return
			}
			result, err := connectedClient.executeCommand(args[0], args[1:])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(result))
		})
		server.httpServer.AddRoute("/"+connectedClient.connection.GetName()+"/command/"+command+"/", func(w http.ResponseWriter, r *http.Request) {
			args := r.URL.Path[len("/"+connectedClient.connection.GetName()+"/command/"):]
			argsSplit := strings.Split(args, " ")
			if len(argsSplit) == 0 {
				http.Error(w, "No command", http.StatusBadRequest)
				return
			}
			result, err := connectedClient.executeCommand(argsSplit[0], argsSplit[1:])
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

	commands := DashboardHelpers.GetCommands(connectedClient.client)
	if commands == nil {
		server.errorLogger.Log("Failed to get commands for connectedClient \"" + connectedClient.connection.GetName() + "\"")
		return
	}
	for command := range commands {
		server.httpServer.RemoveRoute("/" + connectedClient.connection.GetName() + "/command/" + command)
		server.httpServer.RemoveRoute("/" + connectedClient.connection.GetName() + "/command/" + command + "/")
	}
}
