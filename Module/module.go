package Module

import "Systemge/Utilities"

type Module interface {
	Start() error
	Stop() error
	GetCommandHandlers() map[string]CommandHandler
	GetName() string
	GetLogger() *Utilities.Logger
	IsStarted() bool
}

type CommandHandler func([]string) error
