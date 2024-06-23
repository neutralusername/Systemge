package Module

import (
	"Systemge/Application"
	"Systemge/Client"
	"Systemge/HTTPServer"
	"Systemge/Resolution"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
)

type NewApplicationFunc func(*Client.Client, []string) (Application.Application, error)
type NewCompositeApplicationWebsocketFunc func(*Client.Client, []string) (Application.CompositeApplicationWebsocket, error)
type NewCompositeApplicationHTTPFunc func(*Client.Client, []string) (Application.CompositeApplicationHTTP, error)
type NewCompositeApplicationtWebsocketHTTPFunc func(*Client.Client, []string) (Application.CompositeApplicationWebsocketHTTP, error)

type ClientConfig struct {
	Name       string
	LoggerPath string

	ResolverAddress        string
	ResolverNameIndication string
	ResolverTLSCertPath    string

	HTTPPort string
	HTTPCert string
	HTTPKey  string

	WebsocketPattern string
	WebsocketPort    string
	WebsocketCert    string
	WebsocketKey     string
}

func NewClient(clientConfig *ClientConfig, newApplicationFunc NewApplicationFunc, args []string) *Client.Client {
	resolverResolution := Resolution.New("resolver", clientConfig.ResolverAddress, clientConfig.ResolverNameIndication, Utilities.GetFileContent(clientConfig.ResolverTLSCertPath))
	client := Client.New(clientConfig.Name, resolverResolution, Utilities.NewLogger(clientConfig.LoggerPath))
	application, err := newApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	client.SetApplication(application)
	return client
}

func NewCompositeClientHTTP(clientConfig *ClientConfig, newCompositeApplicationHTTPFunc NewCompositeApplicationHTTPFunc, args []string) *Client.Client {
	resolverResolution := Resolution.New("resolver", clientConfig.ResolverAddress, clientConfig.ResolverNameIndication, Utilities.GetFileContent(clientConfig.ResolverTLSCertPath))
	client := Client.New(clientConfig.Name, resolverResolution, Utilities.NewLogger(clientConfig.LoggerPath))
	application, err := newCompositeApplicationHTTPFunc(client, args)
	if err != nil {
		panic(err)
	}
	httpServer := HTTPServer.New(clientConfig.HTTPPort, clientConfig.Name+"HTTP", clientConfig.HTTPCert, clientConfig.HTTPKey, Utilities.NewLogger(clientConfig.LoggerPath), application)
	client.SetApplication(application)
	client.SetHTTPServer(httpServer)
	return client
}

func NewCompositeClientWebsocket(clientConfig *ClientConfig, newCompositeApplicationWebsocketFunc NewCompositeApplicationWebsocketFunc, args []string) *Client.Client {
	resolverResolution := Resolution.New("resolver", clientConfig.ResolverAddress, clientConfig.ResolverNameIndication, Utilities.GetFileContent(clientConfig.ResolverTLSCertPath))
	client := Client.New(clientConfig.Name, resolverResolution, Utilities.NewLogger(clientConfig.LoggerPath))
	application, err := newCompositeApplicationWebsocketFunc(client, args)
	if err != nil {
		panic(err)
	}
	websocketServer := WebsocketServer.New(clientConfig.Name, Utilities.NewLogger(clientConfig.LoggerPath), application)
	websocketServer.SetHTTPServer(HTTPServer.New(clientConfig.WebsocketPort, clientConfig.Name+"WebsocketHandshake", clientConfig.WebsocketCert, clientConfig.WebsocketKey, Utilities.NewLogger(clientConfig.LoggerPath), WebsocketServer.NewHandshakeApplication(clientConfig.WebsocketPattern, websocketServer)))
	client.SetApplication(application)
	client.SetWebsocketServer(websocketServer)
	return client
}

func NewCompositeClientWebsocketHTTP(clientConfig *ClientConfig, newCompositeApplicationtWebsocketHTTPFunc NewCompositeApplicationtWebsocketHTTPFunc, args []string) *Client.Client {
	resolverResolution := Resolution.New("resolver", clientConfig.ResolverAddress, clientConfig.ResolverNameIndication, Utilities.GetFileContent(clientConfig.ResolverTLSCertPath))
	client := Client.New(clientConfig.Name, resolverResolution, Utilities.NewLogger(clientConfig.LoggerPath))
	application, err := newCompositeApplicationtWebsocketHTTPFunc(client, args)
	if err != nil {
		panic(err)
	}
	websocketServer := WebsocketServer.New(clientConfig.Name, Utilities.NewLogger(clientConfig.LoggerPath), application)
	websocketServer.SetHTTPServer(HTTPServer.New(clientConfig.WebsocketPort, clientConfig.Name+"WebsocketHandshake", clientConfig.WebsocketCert, clientConfig.WebsocketKey, Utilities.NewLogger(clientConfig.LoggerPath), WebsocketServer.NewHandshakeApplication(clientConfig.WebsocketPattern, websocketServer)))
	httpServer := HTTPServer.New(clientConfig.HTTPPort, clientConfig.Name+"HTTP", clientConfig.HTTPCert, clientConfig.HTTPKey, Utilities.NewLogger(clientConfig.LoggerPath), application)
	client.SetApplication(application)
	client.SetWebsocketServer(websocketServer)
	client.SetHTTPServer(httpServer)
	return client
}
