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
	acceptHandler         tools.AcceptHandlerWithError[systemge.Connection[D]]

	readerServerAsyncConfig *configs.ReaderAsync
	readerRoutineConfig     *configs.Routine
	readHandler             tools.ReadHandler[D, systemge.Connection[D]]

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
		acceptHandler:         acceptHandler,

		readerServerAsyncConfig: readerServerAsyncConfig,
		readerRoutineConfig:     readerRoutineConfig,
		readHandler:             readHandler,
	}

	accepter, err := serviceAccepter.New(
		listener,
		accepterConfig,
		accepterRoutineConfig,
		func(connection systemge.Connection[D]) error {
			if err := server.acceptHandler(connection); err != nil {
				return err
			}

			reader, err := serviceReader.NewAsync[D](
				connection,
				readerServerAsyncConfig,
				readerRoutineConfig,
				func(d D, c systemge.Connection[D]) {
					server.readHandler(d, c)
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

func (s *Server[D]) GetReadHandler() tools.ReadHandler[D, systemge.Connection[D]] {
	return s.readHandler
}

func (s *Server[D]) GetAcceptHandler() tools.AcceptHandlerWithError[systemge.Connection[D]] {
	return s.acceptHandler
}

func (s *Server[D]) SetAcceptHandler(acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]]) {
	s.acceptHandler = acceptHandler
}

func (s *Server[D]) SetReadHandler(readHandler tools.ReadHandler[D, systemge.Connection[D]]) {
	s.readHandler = readHandler
}
