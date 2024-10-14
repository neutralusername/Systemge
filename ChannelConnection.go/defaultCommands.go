package ChannelConnection

import "github.com/neutralusername/Systemge/Commands"

func (listener *ChannelConnection[T]) GetDefaultCommands() Commands.Handlers {
	return nil
}
