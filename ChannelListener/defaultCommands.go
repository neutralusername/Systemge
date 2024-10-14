package WebsocketListener

import "github.com/neutralusername/Systemge/Commands"

func (listener *ChannelListener[T]) GetDefaultCommands() Commands.Handlers {
	return nil
}
