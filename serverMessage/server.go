package serverMessage

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/server"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.Accepter,
	accepterRoutineConfig *configs.Routine,
	readerServerAsyncConfig *configs.ReaderAsync,
	readerRoutineConfig *configs.Routine,
	asyncMessageHandlers systemge.AsyncMessageHandlers[D],
	syncMessageHandlers systemge.SyncMessageHandlers[D],
	acceptHandler tools.AcceptHandlerWithError[D],
) *server.Server[D] {

}
