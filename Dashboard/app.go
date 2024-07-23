package Dashboard

import (
	"Systemge/Config"
	"Systemge/HTTP"
	"Systemge/Node"
	"net/http"
	"runtime"
	"strings"
)

type App struct {
	nodes               map[string]*Node.Node
	node                *Node.Node
	config              *Config.Dashboard
	started             bool
	httpMessageHandlers map[string]http.HandlerFunc
}

func New(config *Config.Dashboard, nodes ...*Node.Node) *App {
	app := &App{
		nodes:  make(map[string]*Node.Node),
		config: config,
	}

	app.httpMessageHandlers = map[string]http.HandlerFunc{}
	for _, node := range nodes {
		app.nodes[node.GetName()] = node
		if config.AutoStart {
			err := node.Start()
			if err != nil {
				panic(err)
			}
		}
		app.registerNodeHttpHandlers(node)
	}
	_, filePath, _, _ := runtime.Caller(0)
	app.httpMessageHandlers["/"] = HTTP.SendDirectory(filePath[:len(filePath)-len("app.go")] + "frontend")
	return app
}

func (app *App) registerNodeHttpHandlers(node *Node.Node) {
	_, filePath, _, _ := runtime.Caller(0)
	app.httpMessageHandlers["/"+node.GetName()] = func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+node.GetName(), http.FileServer(http.Dir(filePath[:len(filePath)-len("app.go")]+"frontend"))).ServeHTTP(w, r)
	}
	app.httpMessageHandlers["/"+node.GetName()+"/start"] = func(w http.ResponseWriter, r *http.Request) {
		err := node.Start()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
	app.httpMessageHandlers["/"+node.GetName()+"/stop"] = func(w http.ResponseWriter, r *http.Request) {
		err := node.Stop()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write([]byte("OK"))
	}
	app.httpMessageHandlers["/"+node.GetName()+"/command/"] = func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Path[len("/"+node.GetName()+"/command/"):]
		argsSplit := strings.Split(args, " ")
		if len(argsSplit) == 0 {
			http.Error(w, "No command", http.StatusBadRequest)
			return
		}
		result, err := app.nodeCommand(&Command{Name: node.GetName(), Command: argsSplit[0], Args: argsSplit[1:]})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	}
}
