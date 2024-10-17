package SingleMessageServer

import (
	"github.com/neutralusername/systemge/singleRequestServer"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewSingleMessageServerAsync[B any](messageHandlers systemge.AsyncMessageHandlers[B]) *singleRequestServer.SingleRequestServerAsync[*tools.IMessage] {

}
