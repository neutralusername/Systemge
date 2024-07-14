package Module

import "Systemge/Utilities"

type Module interface {
	Start() error
	Stop() error
	GetCustomCommandHandlers() map[string]CustomCommandHandler
	GetName() string
	GetLogger() *Utilities.Logger
}

type CustomCommandHandler func([]string) error
