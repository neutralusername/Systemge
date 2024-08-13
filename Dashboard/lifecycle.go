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
	"github.com/neutralusername/Systemge/Tools"
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
	app.mutex.Lock()
	for _, node := range app.nodes {
		app.registerNodeHttpHandlers(node)
	}
	app.mutex.Unlock()
	_, filePath, _, _ := runtime.Caller(0)
	app.node.AddHttpRoute("/", HTTP.SendDirectory(filePath[:len(filePath)-len("lifecycle.go")]+"frontend"))
	if app.config.AutoStart {
		for _, node := range app.nodes {
			go func() {
				err := node.Start()
				if err != nil {
					if errorLogger := app.node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log("Error starting node \"" + node.GetName() + "\": " + err.Error())
					}
					if mailer := app.node.GetMailer(); mailer != nil {
						err := mailer.Send(Tools.NewMail(nil, "error", "Error starting node \""+node.GetName()+"\": "+err.Error()))
						if err != nil {
							if errorLogger := app.node.GetErrorLogger(); errorLogger != nil {
								errorLogger.Log("Error sending mail: " + err.Error())
							}
						}
					}
				}
			}()
		}
	}
	if app.config.AddDashboardToDashboard {
		app.mutex.Lock()
		app.nodes[node.GetName()] = node
		app.mutex.Unlock()
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

	if app.config.NodeSystemgeClientCounterIntervalMs > 0 {
		go app.nodeSystemgeClientCountersRoutine()
	}
	if app.config.NodeSystemgeClientRateLimitCounterIntervalMs > 0 {
		go app.nodeSystemgeClientRateLimitCountersRoutine()
	}
	if app.config.NodeSystemgeClientConnectionCounterIntervalMs > 0 {
		go app.nodeSystemgeClientConnectionCountersRoutine()
	}
	if app.config.NodeSystemgeClientAsyncMessageCounterIntervalMs > 0 {
		go app.nodeSystemgeClientAsyncMessageCountersRoutine()
	}
	if app.config.NodeSystemgeClientSyncRequestCounterIntervalMs > 0 {
		go app.nodeSystemgeClientSyncRequestCountersRoutine()
	}
	if app.config.NodeSystemgeClientSyncResponseCounterIntervalMs > 0 {
		go app.nodeSystemgeClientSyncResponseCountersRoutine()
	}
	if app.config.NodeSystemgeClientTopicCounterIntervalMs > 0 {
		go app.nodeSystemgeClientTopicCountersRoutine()
	}

	if app.config.NodeSystemgeServerCounterIntervalMs > 0 {
		go app.nodeSystemgeServerCountersRoutine()
	}
	if app.config.NodeSystemgeServerRateLimitCounterIntervalMs > 0 {
		go app.nodeSystemgeServerRateLimitCountersRoutine()
	}
	if app.config.NodeSystemgeServerConnectionCounterIntervalMs > 0 {
		go app.nodeSystemgeServerConnectionCountersRoutine()
	}
	if app.config.NodeSystemgeServerSyncResponseCounterIntervalMs > 0 {
		go app.nodeSystemgeServerSyncResponseCountersRoutine()
	}
	if app.config.NodeSystemgeServerAsyncMessageCounterIntervalMs > 0 {
		go app.nodeSystemgeServerAsyncMessageCountersRoutine()
	}
	if app.config.NodeSystemgeServerSyncRequestCounterIntervalMs > 0 {
		go app.nodeSystemgeServerSyncRequestCountersRoutine()
	}
	if app.config.NodeSystemgeServerTopicCounterIntervalMs > 0 {
		go app.nodeSystemgeServerTopicCountersRoutine()
	}

	if app.config.NodeHTTPCounterIntervalMs > 0 {
		go app.nodeHTTPCountersRoutine()
	}
	if app.config.NodeWebsocketCounterIntervalMs > 0 {
		go app.nodeWebsocketCountersRoutine()
	}
	if app.config.NodeSpawnerCounterIntervalMs > 0 {
		go app.nodeSpawnerCountersRoutine()
	}
	app.mutex.RLock()
	for _, node := range app.nodes {
		if Spawner.ImplementsSpawner(node.GetApplication()) {
			go app.spawnerNodeChangeRoutine(node)
		}
	}
	app.mutex.RUnlock()
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
		go app.node.WebsocketBroadcast(Message.NewAsync("goroutineCount", strconv.Itoa(goroutineCount)))
		if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("goroutine update routine: \"" + strconv.Itoa(goroutineCount) + "\"")
		}
		time.Sleep(time.Duration(app.config.GoroutineUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *App) spawnerNodeChangeRoutine(node *Node.Node) {
	spawner := node.GetApplication().(*Spawner.Spawner)
	for {
		spawnerNodeChange := spawner.GetNextNodeChange()
		if spawnerNodeChange == nil {
			if warningLogger := app.node.GetInternalWarningLogger(); warningLogger != nil {
				warningLogger.Log("Node channel closed for \"" + node.GetName() + "\"")
			}
			return
		}
		app.mutex.Lock()
		if spawnerNodeChange.Added {
			app.nodes[spawnerNodeChange.Node.GetName()] = spawnerNodeChange.Node
			app.registerNodeHttpHandlers(spawnerNodeChange.Node)
			go app.node.WebsocketBroadcast(Message.NewAsync("addNode", Helpers.JsonMarshal(newAddNode(spawnerNodeChange.Node))))

		} else {
			app.unregisterNodeHttpHandlers(spawnerNodeChange.Node)
			delete(app.nodes, spawnerNodeChange.Node.GetName())
			go app.node.WebsocketBroadcast(Message.NewAsync("removeNode", spawnerNodeChange.Node.GetName()))
		}
		app.mutex.Unlock()
	}
}

func (app *App) nodeSpawnerCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Spawner.ImplementsSpawner(node.GetApplication()) {
				spawnerCountersJson := Helpers.JsonMarshal(newNodeSpawnerCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSpawnerCounters", spawnerCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("spawner counter routine: \"" + spawnerCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSpawnerCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeClientCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			systemgeClientCountersJson := Helpers.JsonMarshal(newNodeSystemgeClientCounters(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeClientCounters", systemgeClientCountersJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("systemge client counter routine: \"" + systemgeClientCountersJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeClientCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeClientRateLimitCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			systemgeClientRateLimitCountersJson := Helpers.JsonMarshal(newNodeSystemgeClientRateLimitCounters(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeClientRateLimitCounters", systemgeClientRateLimitCountersJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("systemge client rate limit counter routine: \"" + systemgeClientRateLimitCountersJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeClientRateLimitCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeClientConnectionCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			systemgeClientConnectionCountersJson := Helpers.JsonMarshal(newNodeSystemgeClientConnectionCounters(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeClientConnectionCounters", systemgeClientConnectionCountersJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("systemge client connection counter routine: \"" + systemgeClientConnectionCountersJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeClientConnectionCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeClientAsyncMessageCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			systemgeClientAsyncMessageCountersJson := Helpers.JsonMarshal(newNodeSystemgeClientAsyncMessageCounters(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeClientAsyncMessageCounters", systemgeClientAsyncMessageCountersJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("systemge client async message counter routine: \"" + systemgeClientAsyncMessageCountersJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeClientAsyncMessageCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeClientSyncRequestCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			systemgeClientSyncRequestCountersJson := Helpers.JsonMarshal(newNodeSystemgeClientSyncRequestCounters(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeClientSyncRequestCounters", systemgeClientSyncRequestCountersJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("systemge client sync request counter routine: \"" + systemgeClientSyncRequestCountersJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeClientSyncRequestCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeClientSyncResponseCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			systemgeClientSyncResponseCountersJson := Helpers.JsonMarshal(newNodeSystemgeClientSyncResponseCounters(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeClientSyncResponseCounters", systemgeClientSyncResponseCountersJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("systemge client sync response counter routine: \"" + systemgeClientSyncResponseCountersJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeClientSyncResponseCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeClientTopicCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			systemgeClientTopicCountersJson := Helpers.JsonMarshal(newNodeSystemgeClientTopicCounters(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeClientTopicCounters", systemgeClientTopicCountersJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("systemge client topic counter routine: \"" + systemgeClientTopicCountersJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeClientTopicCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeServerCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeServerComponent(node.GetApplication()) {
				systemgeServerCountersJson := Helpers.JsonMarshal(newNodeSystemgeServerCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeServerCounters", systemgeServerCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge server counter routine: \"" + systemgeServerCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeServerCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeServerRateLimitCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeServerComponent(node.GetApplication()) {
				systemgeServerRateLimitCountersJson := Helpers.JsonMarshal(newNodeSystemgeServerRateLimitCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeServerRateLimitCounters", systemgeServerRateLimitCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge server rate limit counter routine: \"" + systemgeServerRateLimitCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeServerRateLimitCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeServerConnectionCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeServerComponent(node.GetApplication()) {
				systemgeServerConnectionCountersJson := Helpers.JsonMarshal(newNodeSystemgeServerConnectionCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeServerConnectionCounters", systemgeServerConnectionCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge server connection counter routine: \"" + systemgeServerConnectionCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeServerConnectionCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeServerSyncResponseCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeServerComponent(node.GetApplication()) {
				systemgeServerSyncResponseCountersJson := Helpers.JsonMarshal(newNodeSystemgeServerSyncResponseCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeServerSyncResponseCounters", systemgeServerSyncResponseCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge server sync response counter routine: \"" + systemgeServerSyncResponseCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeServerSyncResponseCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeServerAsyncMessageCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeServerComponent(node.GetApplication()) {
				systemgeServerAsyncMessageCountersJson := Helpers.JsonMarshal(newNodeSystemgeServerAsyncMessageCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeServerAsyncMessageCounters", systemgeServerAsyncMessageCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge server async message counter routine: \"" + systemgeServerAsyncMessageCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeServerAsyncMessageCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeServerSyncRequestCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeServerComponent(node.GetApplication()) {
				systemgeServerSyncRequestCountersJson := Helpers.JsonMarshal(newNodeSystemgeServerSyncRequestCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeServerSyncRequestCounters", systemgeServerSyncRequestCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge server sync request counter routine: \"" + systemgeServerSyncRequestCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeServerSyncRequestCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeSystemgeServerTopicCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeServerComponent(node.GetApplication()) {
				systemgeServerTopicCountersJson := Helpers.JsonMarshal(newNodeSystemgeServerTopicCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeServerTopicCounters", systemgeServerTopicCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge server topic counter routine: \"" + systemgeServerTopicCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeServerTopicCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeHTTPCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsHTTPComponent(node.GetApplication()) {
				httpCountersJson := Helpers.JsonMarshal(newHTTPCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeHttpCounters", httpCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("http counter routine: \"" + httpCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeHTTPCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeWebsocketCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsWebsocketComponent(node.GetApplication()) {
				messageCounterJson := Helpers.JsonMarshal(newNodeWebsocketCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeWebsocketCounters", messageCounterJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("websocket message counter routine: \"" + messageCounterJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeWebsocketCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) nodeStatusRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			statusUpdateJson := Helpers.JsonMarshal(newNodeStatus(node))
			go app.node.WebsocketBroadcast(Message.NewAsync("nodeStatus", statusUpdateJson))
			if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("status update routine: \"" + statusUpdateJson + "\"")
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeStatusIntervalMs) * time.Millisecond)
	}
}

func (app *App) heapUpdateRoutine() {
	for app.started {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		heapSize := strconv.FormatUint(memStats.HeapSys, 10)
		go app.node.WebsocketBroadcast(Message.NewAsync("heapStatus", heapSize))
		if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log("heap update routine: \"" + heapSize + "\"")
		}
		time.Sleep(time.Duration(app.config.HeapUpdateIntervalMs) * time.Millisecond)
	}
}
