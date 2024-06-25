package Module

type Module interface {
	Start() error
	Stop() error

	// HandleCustomCommand handles a custom command from the command-line interface
	// The first argument is the command, the second argument is a slice of arguments
	// May return an error if the command is not found or if there is an error executing the command
	GetCustomCommandHandlers() map[string]func([]string) error
}
