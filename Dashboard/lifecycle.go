package Dashboard

import (
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/neutralusername/Systemge/HTTP"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Spawner"
)

func (app *App) registerNodeHttpHandlers(node *Node.Node) {
	_, filePath, _, _ := runtime.Caller(0)

	app.node.AddHttpRoute("/"+node.GetName(), func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/"+node.GetName(), http.FileServer(http.Dir(filePath[:len(filePath)-len("lifecycle.go")]+"frontend"))).ServeHTTP(w, r)
	})
	app.node.AddHttpRoute("/"+node.GetName()+"/command/", func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Path[len("/"+node.GetName()+"/command/"):]
		argsSplit := strings.Split(args, " ")
		if len(argsSplit) == 0 {
			http.Error(w, "No command", http.StatusBadRequest)
			return
		}
		result, err := app.nodeCommand(&Command{Name: node.GetName(), Command: argsSplit[0], Args: argsSplit[1:]})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})
}

func (app *App) unregisterNodeHttpHandlers(node *Node.Node) {
	app.node.RemoveHttpRoute("/" + node.GetName())
	app.node.RemoveHttpRoute("/" + node.GetName() + "/command/")
}

func (app *App) OnStart(node *Node.Node) error {
	app.node = node
	app.started = true
	for _, node := range app.nodes {
		app.registerNodeHttpHandlers(node)
	}
	_, filePath, _, _ := runtime.Caller(0)
	app.node.AddHttpRoute("/", HTTP.SendDirectory(filePath[:len(filePath)-len("lifecycle.go")]+"frontend"))

	if app.config.AddDashboardToDashboard {
		app.nodes[node.GetName()] = node
	}
	if app.config.GoroutineUpdateIntervalMs > 0 {
		go app.goroutineUpdateRoutine()
	}
	if app.config.NodeStatusIntervalMs > 0 {
		go app.nodeStatusRoutine()
	}
	if app.config.HeapUpdateIntervalMs > 0 {
		go app.heapUpdateRoutine()
	}
	if app.config.NodeSystemgeCounterIntervalMs > 0 {
		go app.nodeSystemgeCountersRoutine()
	}
	if app.config.NodeHTTPCounterIntervalMs > 0 {
		go app.nodeHTTPCountersRoutine()
	}
	if app.config.NodeWebsocketCounterIntervalMs > 0 {
		go app.nodeWebsocketCountersRoutine()
	}
	if app.config.NodeBrokerCounterIntervalMs > 0 {
		go app.nodeBrokerCountersRoutine()
	}
	if app.config.NodeResolverCounterIntervalMs > 0 {
		go app.nodeResolverCountersRoutine()
	}
	if app.config.NodeSpawnerCounterIntervalMs > 0 {
		go app.nodeSpawnerCountersRoutine()
	}
	for _, node := range app.nodes {
		if Spawner.ImplementsSpawner(node.GetApplication()) {
			go app.addNodeRoutine(node)
			go app.removeNodeRoutine(node)
		}
	}
	return nil
}

func (app *App) OnStop(node *Node.Node) error {
	app.node = nil
	app.started = false
	return nil
}

func (app *App) goroutineUpdateRoutine() {
	for app.started {
		goroutineCount := runtime.NumGoroutine()
		app.node.WebsocketBroadcast(Message.NewAsync("goroutineCount", app.node.GetName(), strconv.Itoa(goroutineCount)))
		if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("goroutine update routine: \"" + strconv.Itoa(goroutineCount) + "\"")
		}
		time.Sleep(time.Duration(app.config.GoroutineUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *App) addNodeRoutine(node *Node.Node) {
	spawner := node.GetApplication().(*Spawner.Spawner)
	for spawnedNode := range spawner.GetAddNodeChannel() {
		if spawnedNode == nil {
			if warningLogger := app.node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log("Node channel closed for \"" + node.GetName() + "\"")
			}
			return
		}
		app.nodes[spawnedNode.GetName()] = spawnedNode
		app.registerNodeHttpHandlers(spawnedNode)
		app.node.WebsocketBroadcast(Message.NewAsync("nodeStatus", app.node.GetName(), Helpers.JsonMarshal(newNodeStatus(spawnedNode))))
	}
}

func (app *App) removeNodeRoutine(node *Node.Node) {
	spawner := node.GetApplication().(*Spawner.Spawner)
	for removedNode := range spawner.GetRemoveNodeChannel() {
		if removedNode == nil {
			if warningLogger := app.node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log("Node channel closed for \"" + node.GetName() + "\"")
			}
			return
		}
		app.unregisterNodeHttpHandlers(removedNode)
		delete(app.nodes, removedNode.GetName())
		app.node.WebsocketBroadcast(Message.NewAsync("removeNode", app.node.GetName(), removedNode.GetName()))
	}
}

func (app *App) nodeSpawnerCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if Spawner.ImplementsSpawner(node.GetApplication()) {
				spawnerCountersJson := Helpers.JsonMarshal(newNodeSpawnerCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeSpawnerCounters", app.node.GetName(), spawnerCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("spawner counter routine: \"" + spawnerCountersJson + "\"")
				}
			}
		}
		time.Sleep(time.Duration(app.config.NodeSpawnerCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeHTTPCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if Node.ImplementsHTTPComponent(node.GetApplication()) {
				httpCountersJson := Helpers.JsonMarshal(newHTTPCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeHttpCounters", app.node.GetName(), httpCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("http counter routine: \"" + httpCountersJson + "\"")
				}
			}
		}
		time.Sleep(time.Duration(app.config.NodeHTTPCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeResolverCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if Node.ImplementsResolverComponent(node.GetApplication()) {
				resolverCountersJson := Helpers.JsonMarshal(newNodeResolverCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeResolverCounters", app.node.GetName(), resolverCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("resolver counter routine: \"" + resolverCountersJson + "\"")
				}
			}
		}
		time.Sleep(time.Duration(app.config.NodeResolverCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeBrokerCountersRoutine() {
	for app.started {
		for _, node := range app.nodes {
			if Node.ImplementsBrokerComponent(node.GetApplication()) {
				brokerCountersJson := Helpers.JsonMarshal(newNodeBrokerCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeBrokerCounters", app.node.GetName(), brokerCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
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
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				systemgeCountersJson := Helpers.JsonMarshal(newNodeSystemgeCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeCounters", app.node.GetName(), systemgeCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
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
			if Node.ImplementsWebsocketComponent(node.GetApplication()) {
				messageCounterJson := Helpers.JsonMarshal(newNodeWebsocketCounters(node))
				app.node.WebsocketBroadcast(Message.NewAsync("nodeWebsocketCounters", app.node.GetName(), messageCounterJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
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
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
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
		if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("heap update routine: \"" + heapSize + "\"")
		}
		time.Sleep(time.Duration(app.config.HeapUpdateIntervalMs) * time.Millisecond)
	}
}
