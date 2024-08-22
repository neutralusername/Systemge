package Dashboard

import (
	"runtime"
	"strconv"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
)

func (app *DashboardServer) statusUpdateRoutine() {
	for !app.closed {
		app.mutex.RLock()
		for _, client := range app.clients {
			go func() {
				if client.HasStatusFunc {
					response, err := client.connection.SyncRequest(Message.TOPIC_GET_STATUS, "")
					if err != nil {
						if app.errorLogger != nil {
							app.errorLogger.Log("Failed to get status for client \"" + client.Name + "\": " + err.Error())
						}
						return
					}
					status, err := strconv.Atoi(response.GetPayload())
					if err != nil {
						if app.errorLogger != nil {
							app.errorLogger.Log("Failed to parse status for client \"" + client.Name + "\": " + err.Error())
						}
						return
					}
					client.Status = status
					app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(statusUpdate{Name: client.Name, Status: status})))
				}
			}()
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.StatusUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *DashboardServer) metricsUpdateRoutine() {
	for !app.closed {
		app.mutex.RLock()
		for _, client := range app.clients {
			go func() {
				if client.HasMetricsFunc {
					response, err := client.connection.SyncRequest(Message.TOPIC_GET_METRICS, "")
					if err != nil {
						if app.errorLogger != nil {
							app.errorLogger.Log("Failed to get metrics for client \"" + client.Name + "\": " + err.Error())
						}
						return
					}
					metrics, err := unmarshalMetrics(response.GetPayload())
					if err != nil {
						if app.errorLogger != nil {
							app.errorLogger.Log("Failed to parse metrics for client \"" + client.Name + "\": " + err.Error())
						}
						return
					}
					client.Metrics = metrics.Metrics
					app.websocketServer.Broadcast(Message.NewAsync("metricsUpdate", response.GetPayload()))
				}
			}()
		}
		app.mutex.RUnlock()
		time.Sleep(time.Duration(app.config.MetricsUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *DashboardServer) goroutineUpdateRoutine() {
	for !app.closed {
		goroutineCount := runtime.NumGoroutine()
		go app.websocketServer.Broadcast(Message.NewAsync("goroutineCount", strconv.Itoa(goroutineCount)))
		if infoLogger := app.infoLogger; infoLogger != nil {
			infoLogger.Log("goroutine update routine: \"" + strconv.Itoa(goroutineCount) + "\"")
		}
		time.Sleep(time.Duration(app.config.GoroutineUpdateIntervalMs) * time.Millisecond)
	}
}

func (app *DashboardServer) heapUpdateRoutine() {
	for !app.closed {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		heapSize := strconv.FormatUint(memStats.HeapSys, 10)
		go app.websocketServer.Broadcast(Message.NewAsync("heapStatus", heapSize))
		if infoLogger := app.infoLogger; infoLogger != nil {
			infoLogger.Log("heap update routine: \"" + heapSize + "\"")
		}
		time.Sleep(time.Duration(app.config.HeapUpdateIntervalMs) * time.Millisecond)
	}
}
