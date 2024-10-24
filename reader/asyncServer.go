package reader

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

type ReaderAsync[T any] struct {
	connection systemge.Connection[T]

	readRoutine *tools.Routine

	ReadHandler Handler[T]

	// metrics

	SucceededReads atomic.Uint64
	FailedReads    atomic.Uint64
}

func NewAsync[T any](
	connection systemge.Connection[T],
	readerServerAsyncConfig *configs.ReaderAsync,
	routineConfig *configs.Routine,
	readHandler Handler[T],
) (*ReaderAsync[T], error) {

	if connection == nil {
		return nil, errors.New("connection is nil")
	}
	if readerServerAsyncConfig == nil {
		return nil, errors.New("readerServerAsyncConfig is nil")
	}
	if routineConfig == nil {
		return nil, errors.New("routineConfig is nil")
	}
	if readHandler == nil {
		return nil, errors.New("readHandler is nil")
	}

	server := &ReaderAsync[T]{
		ReadHandler: readHandler,
	}

	routine, err := tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			select {
			case <-stopChannel:
				connection.SetReadDeadline(1)
				// routine was stopped
				server.FailedReads.Add(1)
				return

			case <-connection.GetCloseChannel():
				server.readRoutine.Stop()
				// ending routine due to connection close
				server.FailedReads.Add(1)
				return

			case data, ok := <-helpers.ChannelCall(func() (T, error) { return connection.Read(readerServerAsyncConfig.ReadTimeoutNs) }):
				if !ok {
					// do smthg with the error
					server.FailedReads.Add(1)
					return
				}
				server.SucceededReads.Add(1)

				if !readerServerAsyncConfig.HandleReadsConcurrently {
					server.ReadHandler(data, connection)
				} else {
					go server.ReadHandler(data, connection)
				}
			}
		},
		routineConfig,
	)
	if err != nil {
		return nil, err
	}
	server.readRoutine = routine

	return server, nil
}

func (server *ReaderAsync[T]) GetRoutine() *tools.Routine {
	return server.readRoutine
}

func (server *ReaderAsync[T]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("reader_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads": server.SucceededReads.Load(),
			"failedReads":    server.FailedReads.Load(),
		},
	))
	metricsTypes.Merge(server.connection.CheckMetrics())
	return metricsTypes
}
func (server *ReaderAsync[T]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("reader_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads": server.SucceededReads.Swap(0),
			"failedReads":    server.FailedReads.Swap(0),
		},
	))
	metricsTypes.Merge(server.connection.GetMetrics())
	return metricsTypes
}

func (server *ReaderAsync[T]) GetDefaultCommands() tools.CommandHandlers {
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
		return status.ToString(server.readRoutine.GetStatus()), nil
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
	listenerCommands := server.connection.GetDefaultCommands()
	for key, value := range listenerCommands {
		commands["listener_"+key] = value
	}
	return commands
}
