package DashboardServer

import (
	"runtime"
	"strconv"
	"time"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
)

func (app *Server) statusUpdateRoutine() {
	defer app.waitGroup.Done()
	for app.status == Status.STARTED {
		time.Sleep(time.Duration(app.config.StatusUpdateIntervalMs) * time.Millisecond)

		app.mutex.RLock()
		for _, client := range app.clients {
			if DashboardHelpers.HasStatus(client.client.ClientType) {
				go app.clientStatusUpdate(client)
			}
		}
		app.mutex.RUnlock()
	}
}

func (app *Server) metricsUpdateRoutine() {
	defer app.waitGroup.Done()
	for app.status == Status.STARTED {
		time.Sleep(time.Duration(app.config.MetricsUpdateIntervalMs) * time.Millisecond)

		app.mutex.RLock()
		for _, client := range app.clients {
			if DashboardHelpers.HasMetrics(client.client.ClientType) {
				go app.clientMetricsUpdate(client)
			}
		}
		if app.config.Metrics {
			go app.dashboardMetricsUpdate()
		}
		app.mutex.RUnlock()
	}
}

func (app *Server) clientStatusUpdate(client *connectedClient) {
	response, err := client.connection.SyncRequestBlocking(Message.TOPIC_GET_STATUS, "")
	if err != nil {
		if app.errorLogger != nil {
			app.errorLogger.Log("Failed to get status for client \"" + client.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	status, err := strconv.Atoi(response.GetPayload())
	if err != nil {
		if app.errorLogger != nil {
			app.errorLogger.Log("Failed to parse status for client \"" + client.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	client.client.Client.(*DashboardHelpers.CustomServiceClient).Status = status // generalize this
	app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(
		DashboardHelpers.StatusUpdate{
			Name:   client.connection.GetName(),
			Status: status,
		},
	)))
}

func (app *Server) clientMetricsUpdate(client *connectedClient) {
	response, err := client.connection.SyncRequestBlocking(Message.TOPIC_GET_METRICS, "")
	if err != nil {
		if app.errorLogger != nil {
			app.errorLogger.Log("Failed to get metrics for client \"" + client.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	metrics, err := DashboardHelpers.UnmarshalMetrics(response.GetPayload())
	if err != nil {
		if app.errorLogger != nil {
			app.errorLogger.Log("Failed to parse metrics for client \"" + client.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	if metrics.Metrics == nil {
		metrics.Metrics = map[string]uint64{}
	}
	client.client.Client.(*DashboardHelpers.CustomServiceClient).Metrics = metrics.Metrics // generalize this
	metrics.Name = client.connection.GetName()
	app.websocketServer.Broadcast(Message.NewAsync("metricsUpdate", Helpers.JsonMarshal(metrics)))
}

func (app *Server) dashboardMetricsUpdate() {
	systemgeMetrics := app.RetrieveSystemgeMetrics()
	app.websocketServer.Broadcast(Message.NewAsync("dashboardSystemgeMetrics", Helpers.JsonMarshal(systemgeMetrics)))

	websocketMetrics := app.RetrieveWebsocketMetrics()
	app.websocketServer.Broadcast(Message.NewAsync("dashboardWebsocketMetrics", Helpers.JsonMarshal(websocketMetrics)))

	httpMetrics := app.RetrieveHttpMetrics()
	app.websocketServer.Broadcast(Message.NewAsync("dashboardHttpMetrics", Helpers.JsonMarshal(httpMetrics)))
}

func (app *Server) goroutineUpdateRoutine() {
	defer app.waitGroup.Done()
	for app.status == Status.STARTED {
		time.Sleep(time.Duration(app.config.GoroutineUpdateIntervalMs) * time.Millisecond)

		goroutineCount := runtime.NumGoroutine()
		go app.websocketServer.Broadcast(Message.NewAsync("goroutineCount", strconv.Itoa(goroutineCount)))
		if infoLogger := app.infoLogger; infoLogger != nil {
			infoLogger.Log("goroutine update routine: \"" + strconv.Itoa(goroutineCount) + "\"")
		}
	}
}

func (app *Server) heapUpdateRoutine() {
	defer app.waitGroup.Done()
	for app.status == Status.STARTED {
		time.Sleep(time.Duration(app.config.HeapUpdateIntervalMs) * time.Millisecond)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		heapSize := strconv.FormatUint(memStats.HeapSys, 10)
		go app.websocketServer.Broadcast(Message.NewAsync("heapStatus", heapSize))
		if infoLogger := app.infoLogger; infoLogger != nil {
			infoLogger.Log("heap update routine: \"" + heapSize + "\"")
		}
	}
}
