package Dashboard

import (
	"Systemge/Broker"
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
	if app.config.NodeStatusIntervalMs > 0 {
		go app.nodeStatusRoutine()
	}
	if app.config.HeapUpdateIntervalMs > 0 {
		go app.heapUpdateRoutine()
	}
	if app.config.NodeSystemgeCounterIntervalMs > 0 {
		go app.nodeSystemgeCountersRoutine()
	}
	if app.config.NodeWebsocketCounterIntervalMs > 0 {
		go app.nodeWebsocketCountersRoutine()
	}
	if app.config.NodeBrokerCounterIntervalMs > 0 {
		go app.nodeBrokerCountersRoutine()
	}
	return nil
}

func (app *App) OnStop(node *Node.Node) error {
	app.node = nil
	app.started = false
	return nil
}

func (app *App) nodeBrokerCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if Broker.ImplementsBroker(node) {
				brokerCountersJson := Helpers.JsonMarshal(newNodeBrokerCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeBrokerCounters", app.node.GetName(), brokerCountersJson))
				if infoLogger := app.node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log("broker counter routine: \"" + brokerCountersJson + "\"")
				}
			}
		}
		time.Sleep(time.Duration(app.config.NodeBrokerCounterIntervalMs) * time.Millisecond)
	}
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
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeWebsocketCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if node.ImplementsWebsocketComponent() {
				messageCounterJson := Helpers.JsonMarshal(newNodeWebsocketCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeWebsocketCounters", app.node.GetName(), messageCounterJson))
				if infoLogger := app.node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log("websocket message counter routine: \"" + messageCounterJson + "\"")
				}
			}
		}
		time.Sleep(time.Duration(app.config.NodeWebsocketCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeStatusRoutine() {
	for app.started {
		for _, node := range app.nodes {
			statusUpdateJson := Helpers.JsonMarshal(newNodeStatus(node))
			app.node.WebsocketBroadcast(Message.NewAsync("nodeStatus", app.node.GetName(), statusUpdateJson))
			if infoLogger := app.node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log("status update routine: \"" + statusUpdateJson + "\"")
			}
		}
		time.Sleep(time.Duration(app.config.NodeStatusIntervalMs) * time.Millisecond)
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
