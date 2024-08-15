package Module

type ServiceModule interface {
	Service
	Module
}

type Service interface {
	Start() error
	Stop() error
	GetStatus() int
	GetMetrics() map[string]interface{}
}

type Module interface {
	GetName() string
	GetCommandHandlers() map[string]func() (string, error)
}
