package serviceResolver

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceSingleRequest"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandlerWithError[D, systemge.Connection[D]],
	topicData map[string]D,
	deserializeTopic func(D, systemge.Connection[D]) (string, error), // responsible for retrieving the topic
) (*serviceAccepter.Accepter[D], error) {

	return serviceSingleRequest.NewSync(
		listener,
		accepterConfig,
		readerConfig,
		routineConfig,
		true,
		acceptHandler,
		func(closeChannel <-chan struct{}, incomingData D, connection systemge.Connection[D]) (D, error) {
			if err := readHandler(closeChannel, incomingData, connection); err != nil {
				return helpers.GetNilValue(incomingData), err
			}
			topic, err := deserializeTopic(incomingData, connection)
			if err != nil {
				return helpers.GetNilValue(incomingData), err
			}
			outgoingData, ok := topicData[topic]
			if !ok {
				return helpers.GetNilValue(incomingData), errors.New("topic not found")
			}
			return outgoingData, nil
		},
	)
}
