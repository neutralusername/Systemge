package DashboardServer

import (
	"net/http"
	"strings"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
)

func (app *Server) registerModuleHttpHandlers(connectedClient *connectedClient) {
	app.httpServer.AddRoute("/"+connectedClient.connection.GetName(), func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+connectedClient.connection.GetName(), http.FileServer(http.Dir(app.frontendPath))).ServeHTTP(w, r)
	})

	commands, err := DashboardHelpers.GetCommands(connectedClient.client)
	if err != nil {
		app.errorLogger.Log("Failed to get commands for connectedClient \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		return
	}

	for _, command := range commands {
		app.httpServer.AddRoute("/"+connectedClient.connection.GetName()+"/command/"+command, func(w http.ResponseWriter, r *http.Request) {
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
			result, err := connectedClient.ExecuteCommand(args[0], args[1:])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(result))
		})
		app.httpServer.AddRoute("/"+connectedClient.connection.GetName()+"/command/"+command+"/", func(w http.ResponseWriter, r *http.Request) {
			args := r.URL.Path[len("/"+connectedClient.connection.GetName()+"/command/"):]
			argsSplit := strings.Split(args, " ")
			if len(argsSplit) == 0 {
				http.Error(w, "No command", http.StatusBadRequest)
				return
			}
			result, err := connectedClient.ExecuteCommand(argsSplit[0], argsSplit[1:])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(result))
		})
	}
}

func (app *Server) unregisterModuleHttpHandlers(connectedClient *connectedClient) {
	app.httpServer.RemoveRoute("/" + connectedClient.connection.GetName())

	commands, err := DashboardHelpers.GetCommands(connectedClient.client)
	if err != nil {
		app.errorLogger.Log("Failed to get commands for connectedClient \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		return
	}
	for _, command := range commands {
		app.httpServer.RemoveRoute("/" + connectedClient.connection.GetName() + "/command/" + command)
		app.httpServer.RemoveRoute("/" + connectedClient.connection.GetName() + "/command/" + command + "/")
	}
}
