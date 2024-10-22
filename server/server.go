package server

import (
	"github.com/neutralusername/systemge/configs"
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
}

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.Accepter,
	accepterRoutineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],

	readerServerAsyncConfig *configs.ReaderAsync,
	readerRoutineConfig *configs.Routine,
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
) systemge.Server[D] {
	//s := &server[D]{

}
