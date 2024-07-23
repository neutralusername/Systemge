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
	if app.config.NodeSystemgeCountersIntervalMs > 0 {
		go app.nodeSystemgeCountersRoutine()
	}
	return nil
}

func (app *App) OnStop(node *Node.Node) error {
	app.node = nil
	app.started = false
	return nil
}

func (app *App) nodeSystemgeCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if node.ImplementsSystemgeComponent() {
				systemgeCountersJson := Helpers.JsonMarshal(newNodeSystemgeCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeCounters", app.node.GetName(), systemgeCountersJson))
				if infoLogger := app.node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge counter routine: \"" + systemgeCountersJson + "\"")
				}
			}
		}
		time.Sleep(time.Duration(app.config.NodeSystemgeCountersIntervalMs) * time.Millisecond)
	}
}

func (app *App) statusUpdateRoutine() {
	for app.started {
		for _, node := range app.nodes {
			statusUpdateJson := Helpers.JsonMarshal(newNodeStatus(node))
			app.node.WebsocketBroadcast(Message.NewAsync("nodeStatus", app.node.GetName(), statusUpdateJson))
			if infoLogger := app.node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log("status update routine: \"" + statusUpdateJson + "\"")
			}
		}
		time.Sleep(time.Duration(app.config.StatusUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *App) heapUpdateRoutine() {
	for app.started {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		heapSize := strconv.FormatUint(memStats.HeapSys, 10)
		app.node.WebsocketBroadcast(Message.NewAsync("heapStatus", app.node.GetName(), heapSize))
		if infoLogger := app.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log("heap update routine: \"" + heapSize + "\"")
		}
		time.Sleep(time.Duration(app.config.HeapUpdateIntervalMs) * time.Millisecond)
	}
}
