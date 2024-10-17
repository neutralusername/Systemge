package SingleRequestServer

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type SyncSingleRequestServer[B any] struct {
	listener      systemge.Listener[B, systemge.Connection[B]]
	acceptRoutine *tools.Routine

	AcceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]]
	ReadHandler   tools.ReadHandlerWithResult[B, systemge.Connection[B]]

	// metrics

	SucceededCalls atomic.Uint64
	FailedCalls    atomic.Uint64
}

func NewSyncSingleRequestServer[B any](routineConfig *configs.Routine, listener systemge.Listener[B, systemge.Connection[B]], acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]], readHandler tools.ReadHandlerWithResult[B, systemge.Connection[B]]) (*SyncSingleRequestServer[B], error) {

	server := &SyncSingleRequestServer[B]{
		listener:      listener,
		AcceptHandler: acceptHandler,
		ReadHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			connection, err := listener.Accept(acceptTimeoutNs)
			if err != nil {
				server.FailedCalls.Add(1)
				// do smthg with the error
				return
			}
			if err = server.AcceptHandler(connection); err != nil {
				server.FailedCalls.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			object, err := connection.Read(readTimeoutNs)
			if err != nil {
				server.FailedCalls.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			result := server.ReadHandler(object, connection)
			if err = connection.Write(result, writeTimeoutNs); err != nil {
				server.FailedCalls.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			connection.Close()
			server.SucceededCalls.Add(1)
		},
		routineConfig,
	)
	return server, nil
}

func (server *SyncSingleRequestServer[B]) Start() error {
	return server.acceptRoutine.StartRoutine()
}

func (server *SyncSingleRequestServer[B]) Stop() error {
	return server.acceptRoutine.StopRoutine(true)
}

func (server *SyncSingleRequestServer[B]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_sync", tools.NewMetrics(
		map[string]uint64{
			"successes": server.SucceededCalls.Load(),
			"fails":     server.FailedCalls.Load(),
		},
	))
	metricsTypes.Merge(server.listener.CheckMetrics())
	return metricsTypes
}
func (server *SyncSingleRequestServer[B]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_sync", tools.NewMetrics(
		map[string]uint64{
			"successes": server.SucceededCalls.Swap(0),
			"fails":     server.FailedCalls.Swap(0),
		},
	))
	metricsTypes.Merge(server.listener.GetMetrics())
	return metricsTypes
}

func (server *SyncSingleRequestServer[B]) GetDefaultCommands() tools.CommandHandlers {
	commands := tools.CommandHandlers{}
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
		return status.ToString(server.acceptRoutine.GetStatus()), nil
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
	listenerCommands := server.listener.GetDefaultCommands()
	for key, value := range listenerCommands {
		commands["listener_"+key] = value
	}
	return commands
}
