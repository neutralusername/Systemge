package Dashboard

import "Systemge/Node"

func (app *App) OnStart(node *Node.Node) error {
	app.node = node
	return nil
}

func (app *App) OnStop(node *Node.Node) error {
	app.node = nil
	return nil
}
