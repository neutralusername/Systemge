package serviceReader

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type ReaderServerSync[B any] struct {
	connection systemge.Connection[B]

	readRoutine *tools.Routine

	ReadHandler tools.ReadHandlerWithResult[B, systemge.Connection[B]]

	// metrics

	SucceededReads atomic.Uint64
	FailedReads    atomic.Uint64
}

func NewSingleRequestServerSync[B any](
	config *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	connection systemge.Connection[B],
	readHandler tools.ReadHandlerWithResult[B, systemge.Connection[B]],
	handleReadsConcurrently bool,
) (*ReaderServerSync[B], error) {

	server := &ReaderServerSync[B]{
		ReadHandler: readHandler,
	}

	server.readRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			object, err := connection.Read(config.ReadTimeoutNs)
			if err != nil {
				server.FailedReads.Add(1)
				// do smthg with the error
				return
			}

			handleRead := func(object B, connection systemge.Connection[B]) {
				result, err := server.ReadHandler(object, connection)
				if err != nil {
					server.FailedReads.Add(1)
					// do smthg with the error
					return
				}
				err = connection.Write(result, config.WriteTimeoutNs)
				if err != nil {
					server.FailedReads.Add(1)
					// do smthg with the error
					return
				}
				server.SucceededReads.Add(1)
				return
			}

			if !handleReadsConcurrently {
				handleRead(object, connection)
			} else {
				go handleRead(object, connection)
			}
		},
		routineConfig,
	)

	/* 	go func() {
		<-server.connection.GetCloseChannel()
		server.readRoutine.StopRoutine(true)
	}() */

	return server, nil
}

func (server *ReaderServerSync[B]) GetRoutine() *tools.Routine {
	return server.readRoutine
}

func (server *ReaderServerSync[B]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads": server.SucceededReads.Load(),
			"failedReads":    server.FailedReads.Load(),
		},
	))
	metricsTypes.Merge(server.connection.CheckMetrics())
	return metricsTypes
}
func (server *ReaderServerSync[B]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("single_request_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads": server.SucceededReads.Swap(0),
			"failedReads":    server.FailedReads.Swap(0),
		},
	))
	metricsTypes.Merge(server.connection.GetMetrics())
	return metricsTypes
}

func (server *ReaderServerSync[B]) GetDefaultCommands() tools.CommandHandlers {
	commands := tools.CommandHandlers{}
	commands["start"] = func(args []string) (string, error) {
		err := server.GetRoutine().StartRoutine()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		err := server.GetRoutine().StopRoutine(true)
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
