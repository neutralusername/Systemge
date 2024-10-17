package SingleRequestServer

import (
	"encoding/json"

	"github.com/neutralusername/systemge/Commands"
	"github.com/neutralusername/systemge/Metrics"
	"github.com/neutralusername/systemge/status"
)

func (server *AsyncSingleRequestServer) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_syncMessages", Metrics.New(
		map[string]uint64{
			"succeeded_sync_messages": server.CheckSucceededSyncMessages(),
			"failed_sync_messages":    server.CheckFailedSyncMessages(),
		},
	))
	metricsTypes.AddMetrics("single_request_asyncMessages", Metrics.New(
		map[string]uint64{
			"succeeded_async_messages": server.CheckSucceededAsyncMessages(),
			"failed_async_messages":    server.CheckFailedAsyncMessages(),
		},
	))
	metricsTypes.Merge(server.systemgeServer.GetMetrics())
	return metricsTypes
}
func (server *AsyncSingleRequestServer) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_syncMessages", Metrics.New(
		map[string]uint64{
			"succeeded_sync_messages": server.GetSucceededSyncMessages(),
			"failed_sync_messages":    server.GetFailedSyncMessages(),
		},
	))
	metricsTypes.AddMetrics("single_request_asyncMessages", Metrics.New(
		map[string]uint64{
			"succeeded_async_messages": server.GetSucceededAsyncMessages(),
			"failed_async_messages":    server.GetFailedAsyncMessages(),
		},
	))
	metricsTypes.Merge(server.systemgeServer.GetMetrics())
	return metricsTypes
}

func (server *AsyncSingleRequestServer) CheckSucceededAsyncMessages() uint64 {
	return server.succeededAsyncMessages.Load()
}
func (server *AsyncSingleRequestServer) GetSucceededAsyncMessages() uint64 {
	return server.succeededAsyncMessages.Swap(0)
}

func (server *AsyncSingleRequestServer) CheckFailedAsyncMessages() uint64 {
	return server.failedAsyncMessages.Load()
}
func (server *AsyncSingleRequestServer) GetFailedAsyncMessages() uint64 {
	return server.failedAsyncMessages.Swap(0)
}

func (server *AsyncSingleRequestServer) CheckSucceededSyncMessages() uint64 {
	return server.succeededSyncMessages.Load()
}
func (server *AsyncSingleRequestServer) GetSucceededSyncMessages() uint64 {
	return server.succeededSyncMessages.Swap(0)
}

func (server *AsyncSingleRequestServer) CheckFailedSyncMessages() uint64 {
	return server.failedSyncMessages.Load()
}
func (server *AsyncSingleRequestServer) GetFailedSyncMessages() uint64 {
	return server.failedSyncMessages.Swap(0)
}

func (server *AsyncSingleRequestServer) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	commands["start"] = func(args []string) (string, error) {
		err := server.Start()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		err := server.Stop()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return status.ToString(server.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		metrics := server.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := server.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	serverCommands := server.systemgeServer.GetDefaultCommands()
	for key, value := range serverCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
