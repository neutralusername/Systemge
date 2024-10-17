package SingleMessageServer

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/singleRequestServer"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewSingleMessageServerAsync[B any](messageHandlers systemge.AsyncMessageHandlers[B], listener systemge.Listener[B, systemge.Connection[B]], routineConfig *configs.Routine) *singleRequestServer.SingleRequestServerAsync[*tools.IMessage] {

}
