package Resolver

// returns a map of custom command handlers for the command-line interface
func (resolver *Resolver) GetCommandHandlers() map[string]func([]string) error {
	return map[string]func([]string) error{}
}
