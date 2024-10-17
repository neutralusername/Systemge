package singleRequestServer

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type AsyncSingleRequestServer[B any] struct {
	listener      systemge.Listener[B, systemge.Connection[B]]
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]]
	readHandler   tools.ReadHandler[B, systemge.Connection[B]]
	acceptRoutine *tools.Routine

	// metrics

	SucceededCalls atomic.Uint64
	FailedCalls    atomic.Uint64
}

func NewAsyncSingleRequestServer[B any](routineConfig *configs.Routine, acceptTimeoutNs, readTimeoutNs int64, listener systemge.Listener[B, systemge.Connection[B]], acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]], readHandler tools.ReadHandler[B, systemge.Connection[B]]) (*AsyncSingleRequestServer[B], error) {

	server := &AsyncSingleRequestServer[B]{
		listener:      listener,
		acceptHandler: acceptHandler,
		readHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			connection, err := listener.Accept(acceptTimeoutNs)
			if err != nil {
				server.FailedCalls.Add(1)
				// do smthg with the error
				return
			}
			if err = server.acceptHandler(connection); err != nil {
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
			server.readHandler(object, connection)
			connection.Close()
			server.SucceededCalls.Add(1)
		},
		routineConfig,
	)
	return server, nil
}

func (server *AsyncSingleRequestServer[B]) Start() error {
	return server.acceptRoutine.StartRoutine()
}

func (server *AsyncSingleRequestServer[B]) Stop() error {
	return server.acceptRoutine.StopRoutine(true)
}

func (server *AsyncSingleRequestServer[B]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_async", tools.NewMetrics(
		map[string]uint64{
			"successes": server.SucceededCalls.Load(),
			"fails":     server.FailedCalls.Load(),
		},
	))
	metricsTypes.Merge(server.listener.CheckMetrics())
	return metricsTypes
}
func (server *AsyncSingleRequestServer[B]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_async", tools.NewMetrics(
		map[string]uint64{
			"successes": server.SucceededCalls.Swap(0),
			"fails":     server.FailedCalls.Swap(0),
		},
	))
	metricsTypes.Merge(server.listener.GetMetrics())
	return metricsTypes
}

func (server *AsyncSingleRequestServer[B]) GetDefaultCommands() tools.CommandHandlers {
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
