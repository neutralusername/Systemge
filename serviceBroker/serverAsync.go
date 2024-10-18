package serviceBroker

import (
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type Broker[D any] struct {
	mutex         sync.RWMutex
	topics        map[string]map[*subscriber[D]]struct{} // topic -> connection -> struct{}
	subscriptions map[*subscriber[D]]map[string]struct{} // connection -> topic -> struct{}
	accepter      *serviceAccepter.Accepter[D]
}

type subscriber[D any] struct {
	connection  systemge.Connection[D]
	readerAsync *serviceReader.ReaderAsync[D]
}

type HandleMessage[D any] func(
	stopChannel <-chan struct{},
	data D,
	connection systemge.Connection[D],
) (
	subscription bool,
	topic string,
	payload D,
	err error,
)

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterServerConfig *configs.AccepterServer,
	accepterRoutineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	handleAcceptsConcurrently bool,

	readerServerAsyncConfig *configs.ReaderServerAsync,
	readerRoutineConfig *configs.Routine,
	handleReadsConcurrently bool,

	handleMessage HandleMessage[D],
	topics []string,
) (*Broker[D], error) {

	broker := &Broker[D]{
		topics:        make(map[string]map[*subscriber[D]]struct{}),
		subscriptions: make(map[*subscriber[D]]map[string]struct{}),
	}
	for _, topic := range topics {
		broker.topics[topic] = make(map[*subscriber[D]]struct{})
	}

	accepter, err := serviceAccepter.NewAccepterServer(
		listener,
		accepterServerConfig,
		accepterRoutineConfig,
		func(stopChannel <-chan struct{}, connection systemge.Connection[D]) error {
			if err := acceptHandler(stopChannel, connection); err != nil {
				return nil
			}

			readerRoutine, err := serviceReader.NewAsync(
				connection,
				readerServerAsyncConfig,
				readerRoutineConfig,
				nil,
				handleReadsConcurrently,
			)
			if err != nil {
				return err
			}

			subscriber := &subscriber[D]{
				connection:  connection,
				readerAsync: readerRoutine,
			}

			broker.mutex.Lock()
			defer broker.mutex.Unlock()

			broker.subscriptions[subscriber] = make(map[string]struct{})
			return nil
		},
		handleAcceptsConcurrently,
	)
}
