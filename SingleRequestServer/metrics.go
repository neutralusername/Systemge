package SingleRequestServer

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("singleRequestServer_invalidRequests", Metrics.New(
		map[string]uint64{
			"invalid_requests": server.CheckInvalidRequests(),
		},
	))
	metricsTypes.AddMetrics("singleRequestServer_commands", Metrics.New(
		map[string]uint64{
			"succeeded_commands": server.CheckSucceededCommands(),
			"failed_commands":    server.CheckFailedCommands(),
		},
	))
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
func (server *Server) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	metricsTypes.AddMetrics("singleRequestServer_invalidRequests", Metrics.New(
		map[string]uint64{
			"invalid_requests": server.GetInvalidRequests(),
		},
	))
	metricsTypes.AddMetrics("singleRequestServer_commands", Metrics.New(
		map[string]uint64{
			"succeeded_commands": server.GetSucceededCommands(),
			"failed_commands":    server.GetFailedCommands(),
		},
	))
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

func (server *Server) CheckInvalidRequests() uint64 {
	return server.invalidRequests.Load()
}
func (server *Server) GetInvalidRequests() uint64 {
	return server.invalidRequests.Swap(0)
}

func (server *Server) CheckSucceededCommands() uint64 {
	return server.succeededCommands.Load()
}
func (server *Server) GetSucceededCommands() uint64 {
	return server.succeededCommands.Swap(0)
}

func (server *Server) CheckFailedCommands() uint64 {
	return server.failedCommands.Load()
}
func (server *Server) GetFailedCommands() uint64 {
	return server.failedCommands.Swap(0)
}

func (server *Server) CheckSucceededAsyncMessages() uint64 {
	return server.succeededAsyncMessages.Load()
}
func (server *Server) GetSucceededAsyncMessages() uint64 {
	return server.succeededAsyncMessages.Swap(0)
}

func (server *Server) CheckFailedAsyncMessages() uint64 {
	return server.failedAsyncMessages.Load()
}
func (server *Server) GetFailedAsyncMessages() uint64 {
	return server.failedAsyncMessages.Swap(0)
}

func (server *Server) CheckSucceededSyncMessages() uint64 {
	return server.succeededSyncMessages.Load()
}
func (server *Server) GetSucceededSyncMessages() uint64 {
	return server.succeededSyncMessages.Swap(0)
}

func (server *Server) CheckFailedSyncMessages() uint64 {
	return server.failedSyncMessages.Load()
}
func (server *Server) GetFailedSyncMessages() uint64 {
	return server.failedSyncMessages.Swap(0)
}

func (server *Server) GetDefaultCommands() Commands.Handlers {
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
		return Status.ToString(server.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		metrics := server.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Event.New("Failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := server.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Event.New("Failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	serverCommands := server.systemgeServer.GetDefaultCommands()
	for key, value := range serverCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
