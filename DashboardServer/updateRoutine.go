package DashboardServer

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) updateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {
		for _, connectedClient := range server.connectedClients {
			switch connectedClient.page.Type {
			case DashboardHelpers.CLIENT_TYPE_COMMAND:
				if err := server.updateConnectedClientMetrics(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update metrics for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
			case DashboardHelpers.CLIENT_TYPE_CUSTOMSERVICE:
				if err := server.updateConnectedClientStatus(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update status for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientMetrics(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update metrics for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
			case DashboardHelpers.CLIENT_TYPE_SYSTEMGECONNECTION:
				if err := server.updateConnectedClientStatus(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update status for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientMetrics(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update metrics for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientUnprocessedMessageCount(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update unprocessed message count for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientIsProcessingLoopRunning(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update is processing loop running for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
			case DashboardHelpers.CLIENT_TYPE_SYSTEMGESERVER:
				if err := server.updateConnectedClientStatus(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update status for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientMetrics(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update metrics for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientSystemgeConnectionChildren(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update children for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
			}
		}
		if err := server.updateDashboardClientMetrics(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Error.New("Failed to update metrics for dashboard client in updateRoutine", err).Error())
			}
		}
		time.Sleep(time.Duration(server.config.UpdateIntervalMs) * time.Millisecond)
	}
}

func (server *Server) updateConnectedClientStatus(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_GET_STATUS, "")
	if err != nil {
		return Error.New("Failed to execute get status request", err)
	}
	err = connectedClient.page.SetCachedStatus(Helpers.StringToInt(resultPayload))
	if err != nil {
		return Error.New("Failed to set status", err)
	}
	server.mutex.Lock()
	server.dashboardClient.ClientStatuses[connectedClient.connection.GetName()] = Helpers.StringToInt(resultPayload)
	server.mutex.Unlock()

	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DashboardHelpers.DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: map[string]int{
						connectedClient.connection.GetName(): Helpers.StringToInt(resultPayload),
					},
				},
				DashboardHelpers.DASHBOARD_CLIENT_NAME,
			).Marshal(),
		),
	)
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_STATUS: Helpers.StringToInt(resultPayload),
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) updateConnectedClientMetrics(connectedClient *connectedClient) error {
	m := connectedClient.page.GetCachedMetrics()
	if len(m) == 0 {
		return nil
	}
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_GET_METRICS, "")
	if err != nil {
		return Error.New("Failed to execute get metrics request", err)
	}
	metricsTypes, err := Metrics.UnmarshalMetricsTypes(resultPayload)
	if err != nil {
		return Error.New("Failed to unmarshal metrics", err)
	}
	if server.config.MaxMetricsTypes > 0 && len(metricsTypes) > server.config.MaxMetricsTypes {
		return Error.New("Too many metric types", nil)
	}
	for metricsType, metrics := range metricsTypes {
		if metrics == nil {
			return Error.New("Metrics for type "+metricsType+" is nil", nil)
		}
		if server.config.MaxMetricsPerType > 0 && len(metrics.KeyValuePairs) > server.config.MaxMetricsPerType {
			return Error.New("Too many metrics of type "+metricsType, nil)
		}
	}
	for metricsType, metrics := range metricsTypes {
		connectedClient.page.AddCachedMetricsEntry(metricsType, metrics, server.config.MaxEntriesPerMetrics)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_METRICS: DashboardHelpers.NewDashboardMetrics(metricsTypes),
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) updateConnectedClientUnprocessedMessageCount(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_UNHANDLED_MESSAGE_COUNT, "")
	if err != nil {
		return Error.New("Failed to execute unprocessed message count request", err)
	}
	err = connectedClient.page.SetCachedUnprocessedMessageCount(Helpers.StringToUint32(resultPayload))
	if err != nil {
		return Error.New("Failed to set unprocessed message count", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_UNHANDLED_MESSAGE_COUNT: Helpers.StringToUint32(resultPayload),
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) updateConnectedClientIsProcessingLoopRunning(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_IS_MESSAGE_HANDLING_LOOP_STARTED, "")
	if err != nil {
		return Error.New("Failed to execute is processing loop running request", err)
	}
	connectedClient.page.SetCachedIsProcessingLoopRunning(Helpers.StringToBool(resultPayload))
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_IS_MESSAGE_HANDLING_LOOP_STARTED: Helpers.StringToBool(resultPayload),
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) updateConnectedClientSystemgeConnectionChildren(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_GET_SYSTEMGE_CONNECTION_CHILDREN, "")
	if err != nil {
		return Error.New("Failed to execute get children request", err)
	}
	systemgeConnectionChildren := map[string]*DashboardHelpers.SystemgeConnectionChild{}
	if err := json.Unmarshal([]byte(resultPayload), &systemgeConnectionChildren); err != nil {
		return Error.New("Failed to unmarshal children", err)
	}
	err = connectedClient.page.SetCachedSystemgeConnectionChildren(systemgeConnectionChildren)
	if err != nil {
		return Error.New("Failed to set children", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_SYSTEMGE_CONNECTION_CHILDREN: systemgeConnectionChildren,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) updateDashboardClientMetrics() error {
	newMetrics := server.GetMetrics()
	server.addMetricsToDashboardClient(newMetrics)
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DashboardHelpers.DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_METRICS: newMetrics,
				},
				DashboardHelpers.DASHBOARD_CLIENT_NAME,
			).Marshal(),
		),
	)
	return nil
}
