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

func New[D any, O any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterConfig *configs.AccepterServer,
	readerConfig *configs.ReaderServerSync,
	routineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	readHandler tools.ReadHandlerWithError[D, systemge.Connection[D]],
	topicObject map[string]O,
	deserializeTopic func(D, systemge.Connection[D]) (string, error), // responsible for retrieving the topic
	serializeObject func(O) (D, error),
) (*serviceAccepter.AccepterServer[D], error) {

	return serviceSingleRequest.NewSingleRequestServerSync(
		accepterConfig,
		readerConfig,
		routineConfig,
		listener,
		acceptHandler,
		func(closeChannel <-chan struct{}, data D, connection systemge.Connection[D]) (D, error) {
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
