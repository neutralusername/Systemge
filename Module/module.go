package Module

type ServiceModuleWithCommandHandlers interface {
	ServiceModule
	GetCommandHandlers() map[string]func() (string, error)
}

type ServiceModule interface {
	Service
	Module
}

type ModuleWithCommandHandlers interface {
	Module
	GetCommandHandlers() map[string]func() (string, error)
}

type Service interface {
	Start() error
	Stop() error
	GetStatus() int
	GetMetrics() map[string]interface{}
}

type Module interface {
	GetName() string
}
