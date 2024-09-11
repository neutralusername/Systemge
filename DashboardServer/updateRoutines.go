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

func (server *Server) statusUpdateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {
		time.Sleep(time.Duration(server.config.StatusUpdateIntervalMs) * time.Millisecond)

		server.mutex.RLock()
		for _, connectedClient := range server.connectedClients {
			if DashboardHelpers.HasStatus(connectedClient.client) {
				go server.clientStatusUpdate(connectedClient)
			}
		}
		server.mutex.RUnlock()
	}
}

func (server *Server) metricsUpdateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {
		time.Sleep(time.Duration(server.config.MetricsUpdateIntervalMs) * time.Millisecond)

		server.mutex.RLock()
		for _, connectedClient := range server.connectedClients {
			if DashboardHelpers.HasMetrics(connectedClient.client) {
				go server.clientMetricsUpdate(connectedClient)
			}
		}
		if server.config.Metrics {
			go server.dashboardMetricsUpdate()
		}
		server.mutex.RUnlock()
	}
}

func (server *Server) clientStatusUpdate(connectedClient *connectedClient) {
	response, err := connectedClient.connection.SyncRequestBlocking(Message.TOPIC_GET_STATUS, "")
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to get status for connectedClient \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	status, err := strconv.Atoi(response.GetPayload())
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to parse status for connectedClient \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	if !Status.IsValidStatus(status) {
		if server.errorLogger != nil {
			server.errorLogger.Log("Invalid status for connectedClient \"" + connectedClient.connection.GetName() + "\": " + strconv.Itoa(status))
		}
		return
	}
	if err := DashboardHelpers.UpdateStatus(connectedClient.client, status); err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to update status for connectedClient \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	server.websocketServer.Multicast(
		server.getWebsocketClientsOfLocation(""),
		Message.NewAsync("statusUpdate", Helpers.JsonMarshal(
			DashboardHelpers.StatusUpdate{
				Name:   connectedClient.connection.GetName(),
				Status: status,
			},
		)),
	)
}

func (server *Server) clientMetricsUpdate(connectedClient *connectedClient) {
	response, err := connectedClient.connection.SyncRequestBlocking(Message.TOPIC_GET_METRICS, "")
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to get metrics for connectedClient \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	metrics, err := DashboardHelpers.UnmarshalMetrics(response.GetPayload())
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to parse metrics for client \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	if metrics.Metrics == nil {
		metrics.Metrics = map[string]uint64{}
	}
	err = DashboardHelpers.UpdateMetrics(connectedClient.client, metrics.Metrics)
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log("Failed to update metrics for client \"" + connectedClient.connection.GetName() + "\": " + err.Error())
		}
		return
	}
	metrics.Name = connectedClient.connection.GetName()
	server.websocketServer.Multicast(server.getWebsocketClientsOfLocation(connectedClient.connection.GetName()), Message.NewAsync("metricsUpdate", Helpers.JsonMarshal(metrics)))
}

func (server *Server) dashboardMetricsUpdate() {
	systemgeMetrics := server.RetrieveSystemgeMetrics()
	server.websocketServer.Multicast(server.getWebsocketClientsOfLocation(""), Message.NewAsync("dashboardSystemgeMetrics", Helpers.JsonMarshal(systemgeMetrics)))

	websocketMetrics := server.RetrieveWebsocketMetrics()
	server.websocketServer.Multicast(server.getWebsocketClientsOfLocation(""), Message.NewAsync("dashboardWebsocketMetrics", Helpers.JsonMarshal(websocketMetrics)))

	httpMetrics := server.RetrieveHttpMetrics()
	server.websocketServer.Multicast(server.getWebsocketClientsOfLocation(""), Message.NewAsync("dashboardHttpMetrics", Helpers.JsonMarshal(httpMetrics)))
}

func (server *Server) goroutineUpdateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {
		time.Sleep(time.Duration(server.config.GoroutineUpdateIntervalMs) * time.Millisecond)

		goroutineCount := runtime.NumGoroutine()
		go server.websocketServer.Broadcast(Message.NewAsync("goroutineCount", strconv.Itoa(goroutineCount)))
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log("goroutine update routine: \"" + strconv.Itoa(goroutineCount) + "\"")
		}
	}
}

func (server *Server) heapUpdateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {
		time.Sleep(time.Duration(server.config.HeapUpdateIntervalMs) * time.Millisecond)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		heapSize := strconv.FormatUint(memStats.HeapSys, 10)
		go server.websocketServer.Broadcast(Message.NewAsync("heapStatus", heapSize))
		if infoLogger := server.infoLogger; infoLogger != nil {
			infoLogger.Log("heap update routine: \"" + heapSize + "\"")
		}
	}
}

func (server *Server) getWebsocketClientsOfLocation(location string) []string {
	clients := []string{}
	server.mutex.RLock()
	defer server.mutex.RUnlock()
	for websocketId, loc := range server.websocketClientLocations {
		if loc == location {
			clients = append(clients, websocketId)
		}
	}
	return clients
}
