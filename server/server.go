package server

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Server[D any] interface {
}

type basicServer[D any] struct {
	listener              systemge.Listener[D, systemge.Connection[D]]
	accepterConfig        *configs.Accepter
	accepterRoutineConfig *configs.Routine
	acceptHandler         tools.AcceptHandlerWithError[systemge.Connection[D]]

	readerServerAsyncConfig *configs.ReaderAsync
	readerRoutineConfig     *configs.Routine
	readHandler             tools.ReadHandler[D, systemge.Connection[D]]
}
