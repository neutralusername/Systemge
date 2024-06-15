package Client

import "Systemge/Application"

// returns a map of custom command handlers for the command-line interface
func (client *Client) GetCustomCommandHandlers() map[string]Application.CustomCommandHandler {
	return map[string]Application.CustomCommandHandler{}
}
