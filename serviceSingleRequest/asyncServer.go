package serviceSingleRequest

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type SingleRequestServerAsync[D any] struct {
	listener      systemge.Listener[D, systemge.Connection[D]]
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]]
	readHandler   tools.ReadHandler[D, systemge.Connection[D]]
	acceptRoutine *tools.Routine

	// metrics

	SucceededAccepts atomic.Uint64
	FailedAccepts    atomic.Uint64

	SucceededReads atomic.Uint64
	FailedReads    atomic.Uint64
}

func NewSingleRequestServerAsync[D any](
	config *configs.SingleRequestServerAsync,
	routineConfig *configs.Routine,
	listener systemge.Listener[D, systemge.Connection[D]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
) (*SingleRequestServerAsync[D], error) {

	server := &SingleRequestServerAsync[D]{
		listener:      listener,
		acceptHandler: acceptHandler,
		readHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			select {
			case <-stopChannel:
				listener.SetAcceptDeadline(1)
				// routine was stopped
				server.FailedAccepts.Add(1)
				return

			case <-listener.GetStopChannel():
				server.acceptRoutine.Stop()
				// listener was stopped
				server.FailedAccepts.Add(1)
				return

			case connection, ok := <-helpers.ChannelCall(func() (systemge.Connection[D], error) { return listener.Accept(config.AcceptTimeoutNs) }):
				if !ok {
					// do smthg with the error
					server.FailedAccepts.Add(1)
					return
				}
				defer connection.Close()

				if err := server.acceptHandler(connection); err != nil {
					// do smthg with the error
					server.FailedAccepts.Add(1)
					return
				}
				server.SucceededAccepts.Add(1)

				select {
				case <-stopChannel:
					connection.SetReadDeadline(1)
					// routine was stopped
					server.FailedReads.Add(1)
					return

				case <-listener.GetStopChannel():
					server.acceptRoutine.Stop()
					// listener was stopped
					server.FailedReads.Add(1)
					return

				case data, ok := <-helpers.ChannelCall(func() (D, error) { return connection.Read(config.ReadTimeoutNs) }):
					if !ok {
						// do smthg with the error
						server.FailedReads.Add(1)
						return
					}
					server.readHandler(data, connection)
					server.SucceededReads.Add(1)
				}
			}
		},
		routineConfig,
	)
	return server, nil
}

func (server *SingleRequestServerAsync[D]) GetRoutine() *tools.Routine {
	return server.acceptRoutine
}

func (server *SingleRequestServerAsync[D]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_async", tools.NewMetrics(
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
func (server *SingleRequestServerAsync[D]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_async", tools.NewMetrics(
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

func (server *SingleRequestServerAsync[D]) GetDefaultCommands() tools.CommandHandlers {
	commands := tools.CommandHandlers{}
	commands["start"] = func(args []string) (string, error) {
		err := server.GetRoutine().Start()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		err := server.GetRoutine().Stop()
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
