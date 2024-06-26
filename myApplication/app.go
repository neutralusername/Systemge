package myApplication

import "Systemge/Client"

type App struct {
}

func New() Client.Application {
	app := &App{}
	return app
}

func (app *App) OnStart(client *Client.Client) error {
	return nil
}

func (app *App) OnStop(client *Client.Client) error {
	return nil
}
