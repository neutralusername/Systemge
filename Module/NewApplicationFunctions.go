package Module

import (
	"Systemge/Application"
	"Systemge/MessageBrokerClient"
	"Systemge/Utilities"
)

type NewApplicationFunc func(*Utilities.Logger, *MessageBrokerClient.Client) Application.Application
type NewWebsocketApplicationFunc func(*Utilities.Logger, *MessageBrokerClient.Client) Application.WebsocketApplication
