package Application

type CompositeApplicationWebsocket interface {
	Application
	WebsocketApplication
}

type CompositeApplicationHTTP interface {
	Application
	HTTPApplication
}

type CompositeApplicationWebsocketHTTP interface {
	Application
	WebsocketApplication
	HTTPApplication
}
