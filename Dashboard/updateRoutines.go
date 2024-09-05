package Dashboard

import (
	"runtime"
	"strconv"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
)

func (app *DashboardServer) statusUpdateRoutine() {
	defer app.waitGroup.Done()
	for app.status == Status.STARTED {
		time.Sleep(time.Duration(app.config.StatusUpdateIntervalMs) * time.Millisecond)

		app.mutex.RLock()
		for _, client := range app.clients {
			go func() {
				if client.HasStatusFunc {
					response, err := client.connection.SyncRequestBlocking(Message.TOPIC_GET_STATUS, "")
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
	}
}

func (app *DashboardServer) metricsUpdateRoutine() {
	defer app.waitGroup.Done()
	for app.status == Status.STARTED {
		time.Sleep(time.Duration(app.config.MetricsUpdateIntervalMs) * time.Millisecond)

		app.mutex.RLock()
		for _, client := range app.clients {
			go func() {
				if client.HasMetricsFunc {
					response, err := client.connection.SyncRequestBlocking(Message.TOPIC_GET_METRICS, "")
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
					if metrics.Metrics == nil {
						client.Metrics = map[string]uint64{}
					} else {
						client.Metrics = metrics.Metrics
					}
					metrics.Name = client.Name
					app.websocketServer.Broadcast(Message.NewAsync("metricsUpdate", Helpers.JsonMarshal(metrics)))
				}
			}()
		}
		app.mutex.RUnlock()
	}
}

func (app *DashboardServer) goroutineUpdateRoutine() {
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

func (app *DashboardServer) heapUpdateRoutine() {
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
