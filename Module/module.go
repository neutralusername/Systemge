package module

type ServiceModule interface {
	GetName() string
	GetStatus() int
	Start() error
	Stop() error
}

type ApplicationModule interface {
	GetName() string
	GetCommandHandlers() map[string]func() (string, error)
}
