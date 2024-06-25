package Module

type Module interface {
	Start() error
	Stop() error

	GetCustomCommandHandlers() map[string]func([]string) error
}
