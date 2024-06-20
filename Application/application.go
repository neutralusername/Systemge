package Application

import (
	"Systemge/Message"
)

// Application is an interface that defines the methods that an application must implement.
// Every Client must have an Application to Start.
type Application interface {
	// GetAsyncMessageHandlers returns a map of message types and their corresponding handlers.
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler

	// GetSyncMessageHandlers returns a map of message types and their corresponding handlers.
	GetSyncMessageHandlers() map[string]SyncMessageHandler

	// GetCustomCommandHandlers returns a map of custom commands and their corresponding handlers.
	GetCustomCommandHandlers() map[string]CustomCommandHandler

	// OnStart is called when the application is started.
	// Network communication through the Client is possible in this function.
	// If this function returns an error, the application will not be started.
	OnStart() error

	// OnStop is called when the application is stopped.
	// Network communication through the Client is possible in this function.
	// If this function returns an error, the application will not be stopped.
	OnStop() error
}

// AsyncMessageHandler is a function that takes a message and returns an error.
type AsyncMessageHandler func(*Message.Message) error

// SyncMessageHandler is a function that takes a message and returns a response payload and an error.
type SyncMessageHandler func(*Message.Message) (string, error)

// CustomCommandHandler is a function that takes a slice of strings and returns an error.
type CustomCommandHandler func([]string) error
