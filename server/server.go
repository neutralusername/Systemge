package server

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type server[D any] struct {
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
) (systemge.Server[D], error) {
	s := &server[D]{
		listener:              listener,
		accepterConfig:        accepterConfig,
		accepterRoutineConfig: accepterRoutineConfig,
		acceptHandler:         acceptHandler,

		readerServerAsyncConfig: readerServerAsyncConfig,
		readerRoutineConfig:     readerRoutineConfig,
		readHandler:             readHandler,
	}

	accepter, err := serviceAccepter.New[D](
		listener,
		accepterConfig,
		accepterRoutineConfig,
		func(c systemge.Connection[D]) error {
			return s.acceptHandler(c)
		},
	)
	if err != nil {
		return nil, err
	}

	s.accepter = accepter

	return s, nil
}

func (s *server[D]) GetReadHandler() tools.ReadHandler[D, systemge.Connection[D]] {
	return s.readHandler
}

func (s *server[D]) GetAcceptHandler() tools.AcceptHandlerWithError[systemge.Connection[D]] {
	return s.acceptHandler
}

func (s *server[D]) SetAcceptHandler(acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]]) {
	s.acceptHandler = acceptHandler
}

func (s *server[D]) SetReadHandler(readHandler tools.ReadHandler[D, systemge.Connection[D]]) {
	s.readHandler = readHandler
}
