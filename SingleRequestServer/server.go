package SingleRequestServer

import (
	"sync/atomic"

	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type SyncSingleRequestServer[B any] struct {
	listener      systemge.Listener[B, systemge.Connection[B]]
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]]
	readHandler   tools.ReadHandlerWithResult[B, systemge.Connection[B]]
	acceptRoutine *tools.Routine

	// metrics

	succeededAsyncMessages atomic.Uint64
	failedAsyncMessages    atomic.Uint64

	succeededSyncMessages atomic.Uint64
	failedSyncMessages    atomic.Uint64
}

func NewSyncSingleRequestServer[B any](listener systemge.Listener[B, systemge.Connection[B]], acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]], readHandler tools.ReadHandlerWithResult[B, systemge.Connection[B]]) (*AsyncSingleRequestServer[B], error) {

	server := &SyncSingleRequestServer[B]{
		listener:      listener,
		acceptHandler: acceptHandler,
		readHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(func() {
		connection, err := listener.Accept(acceptTimeoutNs)
		if err != nil {
			// do smthg with the error
			return
		}
		err = server.acceptHandler(connection)
		if err != nil {
			// do smthg with the error
			connection.Close()
			return
		}
		object, err := connection.Read(readTimeoutNs)
		if err != nil {
			// do smthg with the error
			connection.Close()
			return
		}
		result := server.readHandler(object, connection)
		err = connection.Write(result, writeTimeoutNs)
		if err != nil {
			// do smthg with the error
			connection.Close()
			return
		}
		connection.Close()

	})
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

	succeededAsyncMessages atomic.Uint64
	failedAsyncMessages    atomic.Uint64

	succeededSyncMessages atomic.Uint64
	failedSyncMessages    atomic.Uint64
}

func NewAsyncSingleRequestServer[B any](listener systemge.Listener[B, systemge.Connection[B]], acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]], readHandler tools.ReadHandler[B, systemge.Connection[B]]) (*AsyncSingleRequestServer[B], error) {

	server := &AsyncSingleRequestServer[B]{
		listener:      listener,
		acceptHandler: acceptHandler,
		readHandler:   readHandler,
	}

	server.acceptRoutine = tools.NewRoutine(func() {
		connection, err := listener.Accept(acceptTimeoutNs)
		if err != nil {
			// do smthg with the error
			return
		}
		err = server.acceptHandler(connection)
		if err != nil {
			// do smthg with the error
			connection.Close()
			return
		}
		object, err := connection.Read(readTimeoutNs)
		if err != nil {
			// do smthg with the error
			connection.Close()
			return
		}
		server.readHandler(object, connection)
		connection.Close()
	})
	return server, nil
}

func (server *AsyncSingleRequestServer[B]) Start() error {
	return server.acceptRoutine.StartRoutine()
}

func (server *AsyncSingleRequestServer[B]) Stop() error {
	return server.acceptRoutine.StopRoutine(true)
}
