package serverMessage

import (
	"github.com/neutralusername/systemge/server"
	"github.com/neutralusername/systemge/systemge"
)

func New[O any](
	asyncMessageHandlers systemge.AsyncMessageHandlers[O],
	syncMessageHandlers systemge.SyncMessageHandlers[O],
) *server.Server[O] {

}
