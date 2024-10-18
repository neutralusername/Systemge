package serviceSingleRequest

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type SingleRequestServerSync[D any] struct {
	listener      systemge.Listener[D, systemge.Connection[D]]
	acceptRoutine *tools.Routine

	AcceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]]
	ReadHandler   tools.ReadHandlerWithResult[D, systemge.Connection[D]]

	// metrics

	SucceededAccepts atomic.Uint64
	FailedAccepts    atomic.Uint64

	SucceededReads atomic.Uint64
	FailedReads    atomic.Uint64
}

func NewSingleRequestServerSync[D any](
	config *configs.SingleRequestServerSync,
	routineConfig *configs.Routine,
	listener systemge.Listener[D, systemge.Connection[D]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandlerWithResult[D, systemge.Connection[D]],
) (*SingleRequestServerSync[D], error) {

	server := &SingleRequestServerSync[D]{
		listener:      listener,
		AcceptHandler: acceptHandler,
		ReadHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {

			connection, err := listener.Accept(config.AcceptTimeoutNs)
			if err != nil {
				server.FailedAccepts.Add(1)
				// do smthg with the error
				return
			}
			defer connection.Close()
			if err = server.AcceptHandler(connection); err != nil {
				server.FailedAccepts.Add(1)
				// do smthg with the error
				return
			}
			server.SucceededAccepts.Add(1)

			object, err := connection.Read(config.ReadTimeoutNs)
			if err != nil {
				server.FailedReads.Add(1)
				// do smthg with the error
				return
			}
			result, err := server.ReadHandler(object, connection)
			if err != nil {
				server.FailedReads.Add(1)
				// do smthg with the error
				return
			}
			if err = connection.Write(result, config.WriteTimeoutNs); err != nil {
				server.FailedReads.Add(1)
				// do smthg with the error
				return
			}
			server.SucceededReads.Add(1)

		},
		routineConfig,
	)
	return server, nil
}

func (server *SingleRequestServerSync[D]) GetRoutine() *tools.Routine {
	return server.acceptRoutine
}

func (server *SingleRequestServerSync[D]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads":   server.SucceededReads.Load(),
			"failedReads":      server.FailedReads.Load(),
			"succeededAccepts": server.SucceededAccepts.Load(),
			"failedAccepts":    server.FailedAccepts.Load(),
		},
	))
	metricsTypes.Merge(server.listener.CheckMetrics())
	return metricsTypes
}
func (server *SingleRequestServerSync[D]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads":   server.SucceededReads.Swap(0),
			"failedReads":      server.FailedReads.Swap(0),
			"succeededAccepts": server.SucceededAccepts.Swap(0),
			"failedAccepts":    server.FailedAccepts.Swap(0),
		},
	))
	metricsTypes.Merge(server.listener.GetMetrics())
	return metricsTypes
}

func (server *SingleRequestServerSync[D]) GetDefaultCommands() tools.CommandHandlers {
	commands := tools.CommandHandlers{}
	commands["start"] = func(args []string) (string, error) {
		err := server.GetRoutine().StartRoutine()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		err := server.GetRoutine().StopRoutine()
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
