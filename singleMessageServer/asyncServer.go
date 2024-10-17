package SingleMessageServer

import (
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/singleRequestServer"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func NewSingleMessageServerAsync[B any](
	config *configs.SingleRequestServerAsync,
	routineConfig *configs.Routine,
	messageHandlers systemge.AsyncMessageHandlers[B],
	listener systemge.Listener[B, systemge.Connection[B]],
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]],
	readHandler tools.ReadHandlerWithError[B, systemge.Connection[B]],
	messageHandler systemge.AsyncMessageHandler[B],
	deserialize func(B) (tools.IMessage, error),
) (*singleRequestServer.SingleRequestServerAsync[B], error) {
	// this is pointless? the basic code with all the tools is just too good
	readHandlerWrapper := func(data B, connection systemge.Connection[B]) {
		if err := readHandler(data, connection); err != nil {
			connection.Close()
		}
		message, err := deserialize(data)
		if err != nil {
			connection.Close()
			return
		}
		// handle message
	}

	return singleRequestServer.NewSingleRequestServerAsync(
		config,
		routineConfig,
		listener,
		acceptHandler,
		readHandlerWrapper,
	)
}
