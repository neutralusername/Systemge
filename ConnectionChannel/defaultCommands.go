package ConnectionChannel

import "github.com/neutralusername/Systemge/Commands"

func (connection *ChannelConnection[T]) GetDefaultCommands() Commands.Handlers {
	return nil
}
