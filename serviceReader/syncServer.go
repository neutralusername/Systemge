package serviceReader

import (
	"encoding/json"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type ReaderServerSync[D any] struct {
	connection systemge.Connection[D]

	readRoutine *tools.Routine

	ReadHandler tools.ReadHandlerWithResult[D, systemge.Connection[D]]

	// metrics

	SucceededReads atomic.Uint64
	FailedReads    atomic.Uint64
}

func NewSingleRequestServerSync[D any](
	config *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	connection systemge.Connection[D],
	readHandler tools.ReadHandlerWithResult[D, systemge.Connection[D]],
	handleReadsConcurrently bool,
) (*ReaderServerSync[D], error) {

	server := &ReaderServerSync[D]{
		ReadHandler: readHandler,
	}

	handleRead := func(object D, connection systemge.Connection[D]) {
		result, err := server.ReadHandler(object, connection)
		if err != nil {
			// do smthg with the error
			server.FailedReads.Add(1)
			return
		}
		err = connection.Write(result, config.WriteTimeoutNs)
		if err != nil {
			// do smthg with the error
			server.FailedReads.Add(1)
			return
		}
		server.SucceededReads.Add(1)
		return
	}

	server.readRoutine = tools.NewRoutine(
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

			case data, ok := <-helpers.ChannelCall(func() (D, error) { return connection.Read(config.ReadTimeoutNs) }):
				if !ok {
					// do smthg with the error
					server.FailedReads.Add(1)
					return
				}
				if !handleReadsConcurrently {
					handleRead(data, connection)
				} else {
					go handleRead(data, connection)
				}
			}
		},
		routineConfig,
	)

	return server, nil
}

func (server *ReaderServerSync[D]) GetRoutine() *tools.Routine {
	return server.readRoutine
}

func (server *ReaderServerSync[D]) CheckMetrics() tools.MetricsTypes {
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
func (server *ReaderServerSync[D]) GetMetrics() tools.MetricsTypes {
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

func (server *ReaderServerSync[D]) GetDefaultCommands() tools.CommandHandlers {
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
