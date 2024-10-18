package serviceAccepter

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type AccepterServer[D any] struct {
	listener      systemge.Listener[D, systemge.Connection[D]]
	acceptRoutine *tools.Routine

	AcceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]]

	// metrics

	SucceededAccepts atomic.Uint64
	FailedAccepts    atomic.Uint64
}

func NewAccepterServer[D any](
	config *configs.AccepterServer,
	routineConfig *configs.Routine,
	listener systemge.Listener[D, systemge.Connection[D]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	handleAcceptsConcurrently bool,
) (*AccepterServer[D], error) {

	server := &AccepterServer[D]{
		listener:      listener,
		AcceptHandler: acceptHandler,
	}

	handleAccept := func(connection systemge.Connection[D]) {
		if err := server.AcceptHandler(connection); err != nil {
			connection.Close()
			// do smthg with the error
			server.FailedAccepts.Add(1)
			return
		}
		server.SucceededAccepts.Add(1)
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

				if !handleAcceptsConcurrently {
					handleAccept(connection)
				} else {
					go handleAccept(connection)
				}
			}
		},
		routineConfig,
	)
	return server, nil
}

func (server *AccepterServer[D]) GetRoutine() *tools.Routine {
	return server.acceptRoutine
}

func (server *AccepterServer[D]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("accepter_server", tools.NewMetrics(
		map[string]uint64{
			"succeededAccepts": server.SucceededAccepts.Load(),
			"failedAccepts":    server.FailedAccepts.Load(),
		},
	))
	metricsTypes.Merge(server.listener.CheckMetrics())
	return metricsTypes
}
func (server *AccepterServer[D]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("accepter_server", tools.NewMetrics(
		map[string]uint64{
			"succeededAccepts": server.SucceededAccepts.Swap(0),
			"failedAccepts":    server.FailedAccepts.Swap(0),
		},
	))
	metricsTypes.Merge(server.listener.GetMetrics())
	return metricsTypes
}

func (server *AccepterServer[D]) GetDefaultCommands() tools.CommandHandlers {
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
