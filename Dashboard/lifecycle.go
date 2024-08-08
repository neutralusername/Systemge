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
	if app.config.NodeSystemgeCounterIntervalMs > 0 {
		go app.nodeSystemgeCountersRoutine()
	}
	if app.config.NodeSystemgeInvalidMessageCounterIntervalMs > 0 {
		go app.newNodeSystemgeInvalidMessageCountersRoutine()
	}
	if app.config.NodeSystemgeIncomingConnectionAttemptsCounterIntervalMs > 0 {
		go app.newNodeSystemgeIncomingConnectionAttemptsCountersRoutine()
	}
	if app.config.NodeSystemgeIncomingSyncResponseCounterIntervalMs > 0 {
		go app.newNodeSystemgeIncomingSyncResponseCountersRoutine()
	}
	if app.config.NodeSystemgeIncomingSyncRequestCounterIntervalMs > 0 {
		go app.newNodeSystemgeIncomingSyncRequestCountersRoutine()
	}
	if app.config.NodeSystemgeIncomingAsyncMessageCounterIntervalMs > 0 {
		go app.newNodeSystemgeIncomingAsyncMessageCountersRoutine()
	}
	if app.config.NodeSystemgeOutgoingConnectionAttemptCounterIntervalMs > 0 {
		go app.newNodeSystemgeOutgoingConnectionAttemptCountersRoutine()
	}
	if app.config.NodeSystemgeOutgoingSyncRequestCounterIntervalMs > 0 {
		go app.newNodeSystemgeOutgoingSyncRequestCountersRoutine()
	}
	if app.config.NodeSystemgeOutgoingAsyncMessageCounterIntervalMs > 0 {
		go app.newNodeSystemgeOutgoingAsyncMessageCountersRoutine()
	}
	if app.config.NodeSystemgeOutgoingSyncResponsesCounterIntervalMs > 0 {
		go app.newNodeSystemgeOutgoingSyncResponsesCountersRoutine()
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

func (app *App) nodeSystemgeCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				systemgeCountersJson := Helpers.JsonMarshal(newNodeSystemgeCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeCounters", systemgeCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("systemge counter routine: \"" + systemgeCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeIncomingSyncResponseCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				incomingSyncResponseCountersJson := Helpers.JsonMarshal(newNodeSystemgeIncomingSyncResponseCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeIncomingSyncResponseCounters", incomingSyncResponseCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("incoming sync response counter routine: \"" + incomingSyncResponseCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeIncomingSyncRequestCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				incomingSyncRequestCountersJson := Helpers.JsonMarshal(newNodeSystemgeIncomingSyncRequestCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeIncomingSyncRequestCounters", incomingSyncRequestCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("incoming sync request counter routine: \"" + incomingSyncRequestCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeIncomingAsyncMessageCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				incomingAsyncMessageCountersJson := Helpers.JsonMarshal(newNodeSystemgeIncomingAsyncMessageCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeIncomingAsyncMessageCounters", incomingAsyncMessageCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("incoming async message counter routine: \"" + incomingAsyncMessageCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeInvalidMessageCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				invalidMessageCountersJson := Helpers.JsonMarshal(newNodeSystemgeInvalidMessageCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeInvalidMessageCounters", invalidMessageCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("invalid message counter routine: \"" + invalidMessageCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeOutgoingSyncRequestCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				outgoingSyncRequestCountersJson := Helpers.JsonMarshal(newNodeSystemgeOutgoingSyncRequestCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeOutgoingSyncRequestCounters", outgoingSyncRequestCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("outgoing sync request counter routine: \"" + outgoingSyncRequestCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeOutgoingAsyncMessageCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				outgoingAsyncMessageCountersJson := Helpers.JsonMarshal(newNodeSystemgeOutgoingAsyncMessageCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeOutgoingAsyncMessageCounters", outgoingAsyncMessageCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("outgoing async message counter routine: \"" + outgoingAsyncMessageCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeIncomingConnectionAttemptsCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				incomingConnectionAttemptsCountersJson := Helpers.JsonMarshal(newNodeSystemgeIncomingConnectionAttemptsCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeIncomingConnectionAttemptsCounters", incomingConnectionAttemptsCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("incoming connection attempts counter routine: \"" + incomingConnectionAttemptsCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeOutgoingConnectionAttemptCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				outgoingConnectionAttemptCountersJson := Helpers.JsonMarshal(newNodeSystemgeOutgoingConnectionAttemptCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeOutgoingConnectionAttemptCounters", outgoingConnectionAttemptCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("outgoing connection attempt counter routine: \"" + outgoingConnectionAttemptCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
	}
}

func (app *App) newNodeSystemgeOutgoingSyncResponsesCountersRoutine() {
	for app.started {
		app.mutex.RLock()
		for _, node := range app.nodes {
			if Node.ImplementsSystemgeComponent(node.GetApplication()) {
				outgoingSyncResponsesCountersJson := Helpers.JsonMarshal(newNodeSystemgeOutgoingSyncResponsesCounters(node))
				go app.node.WebsocketBroadcast(Message.NewAsync("nodeSystemgeOutgoingSyncResponsesCounters", outgoingSyncResponsesCountersJson))
				if infoLogger := app.node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("outgoing sync responses counter routine: \"" + outgoingSyncResponsesCountersJson + "\"")
				}
			}
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.NodeSystemgeCounterIntervalMs) * time.Millisecond)
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
