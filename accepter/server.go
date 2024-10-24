package accepter

import (
	"encoding/json"
	"errors"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Accepter[T any] struct {
	listener      systemge.Listener[T]
	acceptRoutine *tools.Routine

	AcceptHandler HandlerWithError[T]

	// metrics

	SucceededAccepts atomic.Uint64
	FailedAccepts    atomic.Uint64
}

func New[T any](
	listener systemge.Listener[T],
	accepterConfig *configs.Accepter,
	routineConfig *configs.Routine,
	acceptHandler HandlerWithError[T],
) (*Accepter[T], error) {

	if listener == nil {
		return nil, errors.New("listener is nil")
	}
	if accepterConfig == nil {
		return nil, errors.New("accepterConfig is nil")
	}
	if routineConfig == nil {
		return nil, errors.New("routineConfig is nil")
	}
	if acceptHandler == nil {
		return nil, errors.New("acceptHandler is nil")
	}

	server := &Accepter[T]{
		listener:      listener,
		AcceptHandler: acceptHandler,
	}

	handleAccept := func(connection systemge.Connection[T]) {
		if err := server.AcceptHandler(connection); err != nil {
			connection.Close()
			// do smthg with the error
			server.FailedAccepts.Add(1)
			return
		}
		server.SucceededAccepts.Add(1)
	}

	acceptRoutine, err := tools.NewRoutine(
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

			case connection, ok := <-helpers.ChannelCall(func() (systemge.Connection[T], error) { return listener.Accept(accepterConfig.AcceptTimeoutNs) }):
				if !ok {
					// do smthg with the error
					server.FailedAccepts.Add(1)
					return
				}
				if accepterConfig.ConnectionLifetimeNs > 0 {
					tools.NewTimeout(
						accepterConfig.ConnectionLifetimeNs,
						func() {
							connection.Close()
						},
						false,
					)
				}
				if !accepterConfig.HandleAcceptsConcurrently {
					handleAccept(connection)
				} else {
					go handleAccept(connection)
				}
			}
		},
		routineConfig,
	)
	if err != nil {
		return nil, err
	}

	server.acceptRoutine = acceptRoutine

	return server, nil
}

func (server *Accepter[T]) GetRoutine() *tools.Routine {
	return server.acceptRoutine
}

func (server *Accepter[T]) CheckMetrics() tools.MetricsTypes {
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
func (server *Accepter[T]) GetMetrics() tools.MetricsTypes {
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

func (server *Accepter[T]) GetDefaultCommands() tools.CommandHandlers {
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
