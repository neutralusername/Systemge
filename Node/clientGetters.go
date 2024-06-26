package Node

import "Systemge/Utilities"

func (client *Node) GetName() string {
	return client.config.Name
}

func (client *Node) GetApplication() Application {
	return client.application
}

func (client *Node) GetHTTPApplication() HTTPComponent {
	return client.httpApplication
}

func (client *Node) GetWebsocketApplication() WebsocketComponent {
	return client.websocketApplication
}

func (client *Node) GetLogger() *Utilities.Logger {
	return client.logger
}

func (client *Node) GetResolverAddress() string {
	return client.config.ResolverAddress
}

func (client *Node) GetResolverNameIndication() string {
	return client.config.ResolverNameIndication
}

func (client *Node) GetResolverTLSCert() string {
	return client.config.ResolverTLSCert
}
