package SingleRequestServer

import (
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type SyncSingleRequestServer[B any] struct {
	listener      systemge.Listener[B, systemge.Connection[B]]
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]]
	readHandler   tools.ReadHandlerWithResult[B, systemge.Connection[B]]
	acceptRoutine *tools.Routine

	// metrics

	SucceededSyncMessages atomic.Uint64
	FailedSyncMessages    atomic.Uint64
}

func NewSyncSingleRequestServer[B any](routineConfig *configs.Routine, listener systemge.Listener[B, systemge.Connection[B]], acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]], readHandler tools.ReadHandlerWithResult[B, systemge.Connection[B]]) (*SyncSingleRequestServer[B], error) {

	server := &SyncSingleRequestServer[B]{
		listener:      listener,
		acceptHandler: acceptHandler,
		readHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			connection, err := listener.Accept(acceptTimeoutNs)
			if err != nil {
				server.FailedSyncMessages.Add(1)
				// do smthg with the error
				return
			}
			if err = server.acceptHandler(connection); err != nil {
				server.FailedSyncMessages.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			object, err := connection.Read(readTimeoutNs)
			if err != nil {
				server.FailedSyncMessages.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			result := server.readHandler(object, connection)
			if err = connection.Write(result, writeTimeoutNs); err != nil {
				server.FailedSyncMessages.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			connection.Close()
			server.SucceededSyncMessages.Add(1)
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

type AsyncSingleRequestServer[B any] struct {
	listener      systemge.Listener[B, systemge.Connection[B]]
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]]
	readHandler   tools.ReadHandler[B, systemge.Connection[B]]
	acceptRoutine *tools.Routine

	// metrics

	SucceededAsyncMessages atomic.Uint64
	FailedAsyncMessages    atomic.Uint64
}

func NewAsyncSingleRequestServer[B any](routineConfig *configs.Routine, listener systemge.Listener[B, systemge.Connection[B]], acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]], readHandler tools.ReadHandler[B, systemge.Connection[B]]) (*AsyncSingleRequestServer[B], error) {

	server := &AsyncSingleRequestServer[B]{
		listener:      listener,
		acceptHandler: acceptHandler,
		readHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			connection, err := listener.Accept(acceptTimeoutNs)
			if err != nil {
				server.FailedAsyncMessages.Add(1)
				// do smthg with the error
				return
			}
			if err = server.acceptHandler(connection); err != nil {
				server.FailedAsyncMessages.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			object, err := connection.Read(readTimeoutNs)
			if err != nil {
				server.FailedAsyncMessages.Add(1)
				// do smthg with the error
				connection.Close()
				return
			}
			server.readHandler(object, connection)
			connection.Close()
			server.SucceededAsyncMessages.Add(1)
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
