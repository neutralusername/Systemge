package Dashboard

import (
	"Systemge/Helpers"
	"Systemge/Message"
	"Systemge/Node"
	"time"
)

func (app *App) OnStart(node *Node.Node) error {
	app.node = node
	if app.config.StatusUpdateIntervalMs > 0 {
		go app.statusUpdateRoutine()
	}
	return nil
}

func (app *App) OnStop(node *Node.Node) error {
	app.node = nil
	return nil
}

func (app *App) statusUpdateRoutine() {
	for {
		for _, node := range app.nodes {
			app.node.WebsocketBroadcast(Message.NewAsync("nodeStatus", app.node.GetName(), Helpers.JsonMarshal(newNodeStatus(node))))
		}
		time.Sleep(time.Duration(app.config.StatusUpdateIntervalMs) * time.Millisecond)
	}
}
