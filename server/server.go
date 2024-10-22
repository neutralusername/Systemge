package server

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Server[D any] struct {
	listener              systemge.Listener[D, systemge.Connection[D]]
	accepterConfig        *configs.Accepter
	accepterRoutineConfig *configs.Routine

	readerServerAsyncConfig *configs.ReaderAsync
	readerRoutineConfig     *configs.Routine

	ReadHandler   tools.ReadHandler[D, systemge.Connection[D]]
	AcceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]]

	accepter *serviceAccepter.Accepter[D]
}

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.Accepter,
	accepterRoutineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],

	readerServerAsyncConfig *configs.ReaderAsync,
	readerRoutineConfig *configs.Routine,
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
) (*Server[D], error) {
	server := &Server[D]{
		listener:              listener,
		accepterConfig:        accepterConfig,
		accepterRoutineConfig: accepterRoutineConfig,
		AcceptHandler:         acceptHandler,

		readerServerAsyncConfig: readerServerAsyncConfig,
		readerRoutineConfig:     readerRoutineConfig,
		ReadHandler:             readHandler,
	}

	accepter, err := serviceAccepter.New(
		listener,
		accepterConfig,
		accepterRoutineConfig,
		func(connection systemge.Connection[D]) error {
			if err := server.AcceptHandler(connection); err != nil {
				return err
			}

			reader, err := serviceReader.NewAsync[D](
				connection,
				readerServerAsyncConfig,
				readerRoutineConfig,
				func(d D, c systemge.Connection[D]) {
					server.ReadHandler(d, c)
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

func (s *Server[D]) GetAccepter() *serviceAccepter.Accepter[D] {
	return s.accepter
}
