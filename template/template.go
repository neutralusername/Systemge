package main

import (
	"Systemge/Utilities"
	"flag"
)

func main() {
	path := flag.String("path", "./", "path to save the generated files")
	name := flag.String("name", "myApplication", "name of the application")
	http := flag.Bool("http", false, "generate http files")
	websocket := flag.Bool("websocket", false, "generate websocket files")
	flag.Parse()

	Utilities.CreateDirectory(*path + *name + "/")

	GenerateAppFile(*path, *name, *http, *websocket)
	if *http && *websocket {
		GenerateApplicationTemplate(*path, *name)
		GenerateHTTPApplicationTemplate(*path, *name)
		GenerateWebsocketApplicationTemplate(*path, *name)
	} else if *http {
		GenerateApplicationTemplate(*path, *name)
		GenerateHTTPApplicationTemplate(*path, *name)
	} else if *websocket {
		GenerateApplicationTemplate(*path, *name)
		GenerateWebsocketApplicationTemplate(*path, *name)
	} else {
		GenerateApplicationTemplate(*path, *name)
	}
}

func GenerateAppFile(path, name string, http, websocket bool) {
	replacedPackage := Utilities.ReplaceLine(appGo, 0, "package "+name)
	replacedApplicationInterface := ""
	if http && websocket {
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Node.WebsocketHTTPApplication {")
	} else if http {
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Node.HTTPApplication {")
	} else if websocket {
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Node.WebsocketApplication {")
	} else {
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Node.Application {")
	}
	Utilities.OpenFileTruncate(path + name + "/app.go").WriteString(replacedApplicationInterface)
}

func GenerateApplicationTemplate(path, name string) {
	Utilities.OpenFileTruncate(path + name + "/asyncMessageHandlers.go").WriteString(Utilities.ReplaceLine(asyncMessageHandlersGo, 0, "package "+name))
	Utilities.OpenFileTruncate(path + name + "/syncMessageHandlers.go").WriteString(Utilities.ReplaceLine(syncMessageHandlersGo, 0, "package "+name))
	Utilities.OpenFileTruncate(path + name + "/customCommandHandlers.go").WriteString(Utilities.ReplaceLine(customCommandHandlersGo, 0, "package "+name))
}

func GenerateHTTPApplicationTemplate(path, name string) {
	Utilities.OpenFileTruncate(path + name + "/http.go").WriteString(Utilities.ReplaceLine(httpGo, 0, "package "+name))
}

func GenerateWebsocketApplicationTemplate(path, name string) {
	Utilities.OpenFileTruncate(path + name + "/websocket.go").WriteString(Utilities.ReplaceLine(websocketGo, 0, "package "+name))
}

const appGo = `package main

import "Systemge/Node"

type App struct {
}

func New() Node.Application {
	app := &App{}
	return app
}

func (app *App) OnStart(client *Node.Node) error {
	return nil
}

func (app *App) OnStop(client *Node.Node) error {
	return nil
}
`

const asyncMessageHandlersGo = `package main

import (
	"Systemge/Node"
	"Systemge/Message"
)

func (app *App) GetAsyncMessageHandlers() map[string]Node.AsyncMessageHandler {
	return map[string]Node.AsyncMessageHandler{
		"asyncTopic": func(client *Node.Node, message *Message.Message) error {
			return nil
		},
	}
}
`

const syncMessageHandlersGo = `package main

import (
	"Systemge/Node"
	"Systemge/Message"
)

func (app *App) GetSyncMessageHandlers() map[string]Node.SyncMessageHandler {
	return map[string]Node.SyncMessageHandler{
		"syncTopic": func(client *Node.Node, message *Message.Message) (string, error) {
			return "", nil
		},
	}
}
`

const customCommandHandlersGo = `package main

import "Systemge/Node"

func (app *App) GetCustomCommandHandlers() map[string]Node.CustomCommandHandler {
	return map[string]Node.CustomCommandHandler{
		"command": func(client *Node.Node, args []string) error {
			return nil
		},
	}
}
`

const httpGo = `package main

import "Systemge/Node"

func (app *App) GetHTTPRequestHandlers() map[string]Node.HTTPRequestHandler {
	return map[string]Node.HTTPRequestHandler{
		"/": Node.SendHTTPResponseCodeAndBody(200, "Hello, World!"),
	}
}
`

const websocketGo = `package main

import (
	"Systemge/Node"
	"Systemge/Message"
)

func (app *App) GetWebsocketMessageHandlers() map[string]Node.WebsocketMessageHandler {
	return map[string]Node.WebsocketMessageHandler{
		"websocketTopic": func(client *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			return nil
		},
	}
}

func (app *App) OnConnectHandler(client *Node.Node, websocketClient *Node.WebsocketClient) {
	println("websocket client connected")
}

func (app *App) OnDisconnectHandler(client *Node.Node, websocketClient *Node.WebsocketClient) {
	println("websocket client disconnected")
}
`
