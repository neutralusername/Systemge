package DashboardServer

import (
	"net/http"
	"strings"

	"github.com/neutralusername/Systemge/Error"
)

func (app *Server) registerModuleHttpHandlers(connectedClient *connectedClient) {
	app.httpServer.AddRoute("/"+connectedClient.connection.GetName(), func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+connectedClient.connection.GetName(), http.FileServer(http.Dir(app.frontendPath))).ServeHTTP(w, r)
	})

	// todo keep track on which url the client is right now and only propagate required data

	// create a route for each command to avoid command requests for non-existing commands
	app.httpServer.AddRoute("/"+connectedClient.connection.GetName()+"/command", func(w http.ResponseWriter, r *http.Request) {
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
	// create a route for each command to avoid command requests for non-existing commands
	app.httpServer.AddRoute("/"+connectedClient.connection.GetName()+"/command/", func(w http.ResponseWriter, r *http.Request) {
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

func (app *Server) unregisterModuleHttpHandlers(clientName string) {
	app.httpServer.RemoveRoute("/" + clientName)
	app.httpServer.RemoveRoute("/" + clientName + "/command/")
	app.httpServer.RemoveRoute("/" + clientName + "/command")
}
