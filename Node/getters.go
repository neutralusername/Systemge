package Node

import "Systemge/Utilities"

func (node *Node) GetName() string {
	return node.config.Name
}

func (node *Node) GetApplication() Application {
	return node.application
}

func (node *Node) GetHTTPComponent() HTTPComponent {
	return node.httpComponent
}

func (node *Node) GetWebsocketComponent() WebsocketComponent {
	return node.websocketComponent
}

func (node *Node) GetLogger() *Utilities.Logger {
	return node.logger
}

func (node *Node) GetResolverAddress() string {
	return node.application.GetApplicationConfig().ResolverAddress
}

func (node *Node) GetResolverNameIndication() string {
	return node.application.GetApplicationConfig().ResolverNameIndication
}

func (node *Node) GetResolverTLSCert() string {
	return node.application.GetApplicationConfig().ResolverTLSCert
}
