package myApplication

import "Systemge/Client"

func (app *App) GetCustomCommandHandlers() map[string]Client.CustomCommandHandler {
	return map[string]Client.CustomCommandHandler{
		"command": func(client *Client.Client, args []string) error {
			return nil
		},
	}
}
