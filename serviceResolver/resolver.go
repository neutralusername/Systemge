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

func New[B any, O any](
	listener systemge.Listener[B, systemge.Connection[B]],
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[B]],
	readHandler tools.ReadHandlerWithError[B, systemge.Connection[B]],
	topicObject map[string]O,
	deserializeTopic func(B, systemge.Connection[B]) (string, error), // responsible for validating the request and retrieving the topic
	serializeObject func(O) (B, error),
) (*serviceAccepter.AccepterServer[B], error) {

	return serviceSingleRequest.NewSingleRequestServerSync(
		accepterConfig,
		readerConfig,
		routineConfig,
		listener,
		acceptHandler,
		func(closeChannel <-chan struct{}, data B, connection systemge.Connection[B]) (B, error) {
			if err := readHandler(closeChannel, data, connection); err != nil {
				return helpers.GetNilValue(data), err
			}
			topic, err := deserializeTopic(data, connection)
			if err != nil {
				return helpers.GetNilValue(data), err
			}
			object, ok := topicObject[topic]
			if !ok {
				return helpers.GetNilValue(data), errors.New("topic not found")
			}
			return serializeObject(object)
		},
		true,
	)
}
