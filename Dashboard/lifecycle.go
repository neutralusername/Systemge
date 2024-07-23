package Dashboard

import (
	"Systemge/Helpers"
	"Systemge/Message"
	"Systemge/Node"
	"runtime"
	"strconv"
	"time"
)

func (app *App) OnStart(node *Node.Node) error {
	app.node = node
	app.started = true
	if app.config.StatusUpdateIntervalMs > 0 {
		go app.statusUpdateRoutine()
	}
	if app.config.HeapUpdateIntervalMs > 0 {
		go app.heapUpdateRoutine()
	}
	if app.config.MessageCountIntervalMs > 0 {
		go app.nodeCountersRoutine()
	}
	return nil
}

func (app *App) OnStop(node *Node.Node) error {
	app.node = nil
	app.started = false
	return nil
}

func (app *App) nodeCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if node.ImplementsSystemgeComponent() {
				app.node.WebsocketBroadcast(Message.NewAsync("nodeCounters", app.node.GetName(), Helpers.JsonMarshal(newNodeCounters(node))))
			}
		}
		time.Sleep(time.Duration(app.config.MessageCountIntervalMs) * time.Millisecond)
	}
}

func (app *App) statusUpdateRoutine() {
	for app.started {
		for _, node := range app.nodes {
			app.node.WebsocketBroadcast(Message.NewAsync("nodeStatus", app.node.GetName(), Helpers.JsonMarshal(newNodeStatus(node))))
		}
		time.Sleep(time.Duration(app.config.StatusUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *App) heapUpdateRoutine() {
	for app.started {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		app.node.WebsocketBroadcast(Message.NewAsync("heapStatus", app.node.GetName(), strconv.FormatUint(memStats.HeapAlloc, 10)))
		time.Sleep(time.Duration(app.config.HeapUpdateIntervalMs) * time.Millisecond)
	}
}
