package serverMessage

import (
	"github.com/neutralusername/systemge/server"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func New[O any](
	asyncMessageHandlers systemge.AsyncMessageHandlers[O],
	syncMessageHandlers systemge.SyncMessageHandlers[O],
	acceptHandler tools.AcceptHandlerWithError[O],
) *server.Server[O] {

}
