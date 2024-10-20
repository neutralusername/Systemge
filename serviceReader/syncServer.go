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

type ReaderSync[D any] struct {
	connection systemge.Connection[D]

	readRoutine *tools.Routine

	ReadHandler tools.ReadHandlerWithResult[D, systemge.Connection[D]]

	// metrics

	SucceededReads atomic.Uint64
	FailedReads    atomic.Uint64

	SucceededWrites atomic.Uint64
	FailedWrites    atomic.Uint64
}

func NewSync[D any](
	connection systemge.Connection[D],
	readerServerSyncConfig *configs.ReaderSync,
	routineConfig *configs.Routine,
	readHandler tools.ReadHandlerWithResult[D, systemge.Connection[D]],
) (*ReaderSync[D], error) {

	server := &ReaderSync[D]{
		ReadHandler: readHandler,
	}

	handleRead := func(stopChannel <-chan struct{}, object D, connection systemge.Connection[D]) {
		result, err := server.ReadHandler(object, connection)
		if err != nil {
			// do smthg with the error
			server.FailedReads.Add(1)
			return
		}
		server.SucceededReads.Add(1)

		select {
		case <-server.readRoutine.GetStopChannel():
			connection.SetWriteDeadline(1)
			// routine was stopped
			return

		case <-connection.GetCloseChannel():
			// ending routine due to connection close
			return

		case <-helpers.ChannelCall(func() (error, error) {
			if err := connection.Write(result, readerServerSyncConfig.WriteTimeoutNs); err != nil {
				// do smthg with the error
				server.FailedWrites.Add(1)
				return err, err
			}
			return nil, nil
		}):
			server.SucceededWrites.Add(1)
			return
		}
	}

	server.readRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			select {
			case <-stopChannel:
				connection.SetReadDeadline(1)
				// routine was stopped
				return

			case <-connection.GetCloseChannel():
				server.readRoutine.Stop()
				// ending routine due to connection close
				return

			case data, ok := <-helpers.ChannelCall(func() (D, error) { return connection.Read(readerServerSyncConfig.ReadTimeoutNs) }):
				if !ok {
					// do smthg with the error
					server.FailedReads.Add(1)
					return
				}
				if !readerServerSyncConfig.HandleReadsConcurrently {
					handleRead(stopChannel, data, connection)
				} else {
					go handleRead(stopChannel, data, connection)
				}
			}
		},
		routineConfig,
	)

	return server, nil
}

func (server *ReaderSync[D]) GetRoutine() *tools.Routine {
	return server.readRoutine
}

func (server *ReaderSync[D]) CheckMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("reader_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads":  server.SucceededReads.Load(),
			"failedReads":     server.FailedReads.Load(),
			"succeededWrites": server.SucceededWrites.Load(),
			"failedWrites":    server.FailedWrites.Load(),
		},
	))
	metricsTypes.Merge(server.connection.CheckMetrics())
	return metricsTypes
}
func (server *ReaderSync[D]) GetMetrics() tools.MetricsTypes {
	metricsTypes := tools.NewMetricsTypes()
	metricsTypes.AddMetrics("reader_server_sync", tools.NewMetrics(
		map[string]uint64{
			"succeededReads":  server.SucceededReads.Swap(0),
			"failedReads":     server.FailedReads.Swap(0),
			"succeededWrites": server.SucceededWrites.Swap(0),
			"failedWrites":    server.FailedWrites.Swap(0),
		},
	))
	metricsTypes.Merge(server.connection.GetMetrics())
	return metricsTypes
}

func (server *ReaderSync[D]) GetDefaultCommands() tools.CommandHandlers {
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
