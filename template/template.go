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

	println(*http, *websocket)

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
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Client.WebsocketHTTPApplication {")
	} else if http {
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Client.HTTPApplication {")
	} else if websocket {
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Client.WebsocketApplication {")
	} else {
		replacedApplicationInterface = Utilities.ReplaceLine(replacedPackage, 7, "func New() Client.Application {")
	}
	Utilities.OpenFile(path + name + "/app.go").WriteString(replacedApplicationInterface)

}

func GenerateApplicationTemplate(path, name string) {
	Utilities.OpenFile(path + name + "/asyncMessageHandlers.go").WriteString(Utilities.ReplaceLine(asyncMessageHandlersGo, 0, "package "+name))
	Utilities.OpenFile(path + name + "/syncMessageHandlers.go").WriteString(Utilities.ReplaceLine(syncMessageHandlersGo, 0, "package "+name))
	Utilities.OpenFile(path + name + "/customCommandHandlers.go").WriteString(Utilities.ReplaceLine(customCommandHandlersGo, 0, "package "+name))
}

func GenerateHTTPApplicationTemplate(path, name string) {
	Utilities.OpenFile(path + name + "/http.go").WriteString(Utilities.ReplaceLine(httpGo, 0, "package "+name))
}

func GenerateWebsocketApplicationTemplate(path, name string) {
	Utilities.OpenFile(path + name + "/websocket.go").WriteString(Utilities.ReplaceLine(websocketGo, 0, "package "+name))
}

const appGo = `package main

import "Systemge/Client"

type App struct {
}

func New() Client.Application {
	app := &App{}
	return app
}

func (app *App) OnStart(client *Client.Client) error {
	return nil
}

func (app *App) OnStop(client *Client.Client) error {
	return nil
}
`

const asyncMessageHandlersGo = `package main

import (
	"Systemge/Client"
	"Systemge/Message"
)

func (app *App) GetAsyncMessageHandlers() map[string]Client.AsyncMessageHandler {
	return map[string]Client.AsyncMessageHandler{
		"topic": func(client *Client.Client, message *Message.Message) error {
			return nil
		},
	}
}
`

const syncMessageHandlersGo = `package main

import (
	"Systemge/Client"
	"Systemge/Message"
)

func (app *App) GetSyncMessageHandlers() map[string]Client.SyncMessageHandler {
	return map[string]Client.SyncMessageHandler{
		"topic": func(client *Client.Client, message *Message.Message) (string, error) {
			return "", nil
		},
	}
}
`

const customCommandHandlersGo = `package main

import "Systemge/Client"

func (app *App) GetCustomCommandHandlers() map[string]Client.CustomCommandHandler {
	return map[string]Client.CustomCommandHandler{
		"command": func(client *Client.Client, args []string) error {
			return nil
		},
	}
}
`

const httpGo = `package main

import "Systemge/Client"

func (app *App) GetHTTPRequestHandlers() map[string]Client.HTTPRequestHandler {
	return map[string]Client.HTTPRequestHandler{
		"/": Client.SendHTTPResponseCodeAndBody(200, "Hello, World!"),
	}
}
`

const websocketGo = `package main

import (
	"Systemge/Client"
	"Systemge/Message"
)

func (app *App) GetWebsocketMessageHandlers() map[string]Client.WebsocketMessageHandler {
	return map[string]Client.WebsocketMessageHandler{
		"topic": func(client *Client.Client, websocketClient *Client.WebsocketClient, message *Message.Message) error {
			return nil
		},
	}
}

func (app *App) OnConnectHandler(client *Client.Client, websocketClient *Client.WebsocketClient) {
	println("websocket client connected")
}

func (app *App) OnDisconnectHandler(client *Client.Client, websocketClient *Client.WebsocketClient) {
	println("websocket client disconnected")
}
`
