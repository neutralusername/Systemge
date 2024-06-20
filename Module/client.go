package Module

import (
	"Systemge/Application"
	"Systemge/Client"
	"Systemge/HTTP"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
)

type NewApplicationFunc func(*Client.Client, []string) (Application.Application, error)
type NewWebsocketCompositeApplicationFunc func(*Client.Client, []string) (Application.CompositeApplicationWebsocket, error)
type NewHTTPCompositeApplicationFunc func(*Client.Client, []string) (Application.CompositeApplicationHTTP, error)
type NewWebsocketHTTPCompositeApplicationFunc func(*Client.Client, []string) (Application.CompositeApplicationWebsocketHTTP, error)

func NewClient(name string, resolverAddress string, loggerPath string, newApplicationFunc NewApplicationFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	client.SetApplication(application)
	return client
}

func NewCompositeClientHTTP(name string, resolverAddress string, loggerPath string, httpPort string, httpTlsCert string, httpTlsKey string, newHTTPCompositeApplicationFunc NewHTTPCompositeApplicationFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newHTTPCompositeApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	httpServer := HTTP.New(httpPort, name+"HTTP", httpTlsCert, httpTlsKey, Utilities.NewLogger(loggerPath), application)
	client.SetApplication(application)
	client.SetHTTPServer(httpServer)
	return client
}

func NewCompositeClientWebsocket(name string, resolverAddress string, loggerPath string, websocketPattern string, websocketPort string, websocketTlsCert string, websocketTlsKey string, newWebsocketCompositeApplicationFunc NewWebsocketCompositeApplicationFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newWebsocketCompositeApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	websocketServer := WebsocketServer.New(name, Utilities.NewLogger(loggerPath))
	websocketServer.SetHTTPServer(HTTP.New(websocketPort, name+"HTTP", websocketTlsCert, websocketTlsKey, Utilities.NewLogger(loggerPath), WebsocketServer.NewHandshakeApplication(websocketPattern, websocketServer)))
	websocketServer.SetWebsocketApplication(application)
	client.SetApplication(application)
	client.SetWebsocketServer(websocketServer)
	return client
}

func NewCompositeClientWebsocketHTTP(name string, resolverAddress string, loggerPath string, websocketPattern string, websocketPort string, websocketTlsCert string, websocketTlsKey string, httpPort string, httpTlsCert string, httpTlsKey string, newWebsocketHTTPCompositeApplicationFunc NewWebsocketHTTPCompositeApplicationFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newWebsocketHTTPCompositeApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	websocketServer := WebsocketServer.New(name, Utilities.NewLogger(loggerPath))
	websocketServer.SetHTTPServer(HTTP.New(websocketPort, name+"HTTP", websocketTlsCert, websocketTlsKey, Utilities.NewLogger(loggerPath), WebsocketServer.NewHandshakeApplication(websocketPattern, websocketServer)))
	httpServer := HTTP.New(httpPort, name+"HTTP", httpTlsCert, httpTlsKey, Utilities.NewLogger(loggerPath), application)
	websocketServer.SetWebsocketApplication(application)
	client.SetApplication(application)
	client.SetWebsocketServer(websocketServer)
	client.SetHTTPServer(httpServer)
	return client
}
