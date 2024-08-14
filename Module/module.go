package Module

type ServiceModule interface {
	GetName() string
	GetStatus() int
	GetMetrics() map[string]interface{}
	GetCommandHandlers() map[string]func() (string, error)
	Start() error
	Stop() error
}

type ApplicationModule interface {
	GetName() string
	GetCommandHandlers() map[string]func() (string, error)
}
