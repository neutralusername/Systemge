package server

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
)

type Server[T any] struct {
	listener              systemge.Listener[T]
	accepterConfig        *configs.Accepter
	accepterRoutineConfig *configs.Routine

	readerServerAsyncConfig *configs.ReaderAsync
	readerRoutineConfig     *configs.Routine

	ReadHandler   systemge.ReadHandler[T]
	AcceptHandler systemge.AcceptHandlerWithError[T]

	accepter *serviceAccepter.Accepter[T]
}

func New[T any](
	listener systemge.Listener[T],
	accepterConfig *configs.Accepter,
	accepterRoutineConfig *configs.Routine,
	readerServerAsyncConfig *configs.ReaderAsync,
	readerRoutineConfig *configs.Routine,
	acceptHandler systemge.AcceptHandlerWithError[T],
	readHandler systemge.ReadHandler[T],
) (*Server[T], error) {
	server := &Server[T]{
		listener:              listener,
		accepterConfig:        accepterConfig,
		accepterRoutineConfig: accepterRoutineConfig,

		readerServerAsyncConfig: readerServerAsyncConfig,
		readerRoutineConfig:     readerRoutineConfig,

		AcceptHandler: acceptHandler,
		ReadHandler:   readHandler,
	}

	accepter, err := serviceAccepter.New(
		listener,
		accepterConfig,
		accepterRoutineConfig,
		func(connection systemge.Connection[T]) error {
			if err := server.AcceptHandler(connection); err != nil {
				return err
			}

			reader, err := serviceReader.NewAsync[T](
				connection,
				readerServerAsyncConfig,
				readerRoutineConfig,
				func(data T, connection systemge.Connection[T]) {
					server.ReadHandler(data, connection)
				},
			)
			if err != nil {
				return err
			}
			if err := reader.GetRoutine().Start(); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	server.accepter = accepter

	return server, nil
}

func (server *Server[T]) GetAccepter() *serviceAccepter.Accepter[T] {
	return server.accepter
}
