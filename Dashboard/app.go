package Dashboard

/*
func (app *DashboardServer) registerModuleHttpHandlers(module Module.Module) {
	_, filePath, _, _ := runtime.Caller(0)

	app.httpServer.AddRoute("/"+module.GetName(), func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+module.GetName(), http.FileServer(http.Dir(filePath[:len(filePath)-len("lifecycle.go")]+"frontend"))).ServeHTTP(w, r)
	})
	app.httpServer.AddRoute("/"+module.GetName()+"/command/", func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Path[len("/"+module.GetName()+"/command/"):]
		argsSplit := strings.Split(args, " ")
		if len(argsSplit) == 0 {
			http.Error(w, "No command", http.StatusBadRequest)
			return
		}
		result, err := app.nodeCommand(&Command{ModuleName: module.GetName(), Command: argsSplit[0], Args: argsSplit[1:]})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})
}

func (app *DashboardServer) unregisterNodeHttpHandlers(module Module.Module) {
	app.httpServer.RemoveRoute("/" + module.GetName())
	app.httpServer.RemoveRoute("/" + module.GetName() + "/command/")
}
*/
