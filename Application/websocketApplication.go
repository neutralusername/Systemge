package Application

import (
	"Systemge/Message"
	"Systemge/WebsocketClient"
)

type WebsocketApplication interface {
	// GetWebsocketMessageHandlers returns a map of message types and their corresponding handlers.
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler

	// OnConnectHandler is called when a new websocket client connects.
	// The websocket client is passed as an argument.
	// Network communication through the Client is possible in this function.
	// This function is intended for Authentication, Authorization, and Initialization.
	OnConnectHandler(*WebsocketClient.Client)

	// OnDisconnectHandler is called when a websocket client disconnects.
	// The websocket client is passed as an argument.
	// Network communication through the Client is possible in this function.
	// Communication with the disconnected client is no longer possible.
	// This function is intended for Cleanup and Logging.
	OnDisconnectHandler(*WebsocketClient.Client)
}

// WebsocketMessageHandler is a function that takes a websocket client, which sent the message, and the message itself and returns an error.
type WebsocketMessageHandler func(*WebsocketClient.Client, *Message.Message) error
