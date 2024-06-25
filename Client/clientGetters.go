package Client

import "Systemge/Utilities"

func (client *Client) GetName() string {
	return client.config.Name
}

func (client *Client) GetApplication() Application {
	return client.application
}

func (client *Client) GetHTTPApplication() HTTPApplication {
	return client.httpApplication
}

func (client *Client) GetWebsocketApplication() WebsocketApplication {
	return client.websocketApplication
}

func (client *Client) GetLogger() *Utilities.Logger {
	return client.logger
}

func (client *Client) GetResolverAddress() string {
	return client.config.ResolverAddress
}

func (client *Client) GetResolverNameIndication() string {
	return client.config.ResolverNameIndication
}

func (client *Client) GetResolverTLSCert() string {
	return client.config.ResolverTLSCert
}
