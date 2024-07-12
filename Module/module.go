package Module

import "Systemge/Utilities"

type Module interface {
	Start() error
	Stop() error
	GetCustomCommandHandlers() map[string]func([]string) error
	GetName() string
	GetLogger() *Utilities.Logger
}
