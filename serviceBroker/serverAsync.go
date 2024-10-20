package serviceBroker

import (
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type BrokerAsync[D any] struct {
	mutex       sync.RWMutex
	topics      map[string]map[*subscriber[D]]struct{} // topic -> connection -> struct{}
	subscribers map[systemge.Connection[D]]*subscriber[D]
	accepter    *serviceAccepter.Accepter[D]
}

type subscriber[D any] struct {
	connection    systemge.Connection[D]
	readerAsync   *serviceReader.ReaderAsync[D]
	subscriptions map[string]struct{}
}

type HandleMessageAsync[D any] func(
	data D,
	connection systemge.Connection[D],
) (
	subscription bool,
	topic string,
	payload D,
	err error,
)

func NewAsync[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterServerConfig *configs.AccepterServer,
	accepterRoutineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	handleAcceptsConcurrently bool,

	readerServerAsyncConfig *configs.ReaderServerAsync,
	readerRoutineConfig *configs.Routine,
	handleReadsConcurrently bool,

	propagateTimeoutNs int64,
	handleMessage HandleMessageAsync[D],
	topics []string,
) (*BrokerAsync[D], error) {

	broker := &BrokerAsync[D]{
		topics:      make(map[string]map[*subscriber[D]]struct{}),
		subscribers: make(map[systemge.Connection[D]]*subscriber[D]),
	}
	for _, topic := range topics {
		broker.topics[topic] = make(map[*subscriber[D]]struct{})
	}

	accepter, err := serviceAccepter.NewAccepterServer(
		listener,
		accepterServerConfig,
		accepterRoutineConfig,
		handleAcceptsConcurrently,
		func(connection systemge.Connection[D]) error {
			if err := acceptHandler(connection); err != nil {
				return nil
			}

			readerRoutine, err := serviceReader.NewAsync(
				connection,
				readerServerAsyncConfig,
				readerRoutineConfig,
				handleReadsConcurrently,
				func(data D, connection systemge.Connection[D]) {
					subscription, topic, payload, err := handleMessage(data, connection)
					if err != nil {
						return
					}

					if subscription {
						broker.mutex.Lock()
						defer broker.mutex.Unlock()

						subscriber, ok := broker.subscribers[connection]
						if !ok {
							return
						}

						if _, ok := subscriber.subscriptions[topic]; !ok {
							subscriber.subscriptions[topic] = struct{}{}
						} else {
							delete(subscriber.subscriptions, topic)
						}
					} else {
						broker.mutex.RLock()
						defer broker.mutex.RUnlock()

						subscribers, ok := broker.topics[topic]
						if !ok {
							return
						}

						for subscriber := range subscribers {
							if subscriber.connection == connection {
								continue
							}
							go subscriber.connection.Write(payload, propagateTimeoutNs)
						}
					}
				},
			)
			if err != nil {
				return err
			}

			subscriber := &subscriber[D]{
				connection:    connection,
				readerAsync:   readerRoutine,
				subscriptions: make(map[string]struct{}),
			}

			broker.mutex.Lock()
			defer broker.mutex.Unlock()

			broker.subscribers[connection] = subscriber

			go func() {
				select {
				case <-connection.GetCloseChannel():
				case <-broker.accepter.GetRoutine().GetStopChannel():
				}

				broker.mutex.Lock()
				defer broker.mutex.Unlock()

				subscriber := broker.subscribers[connection]
				delete(broker.subscribers, connection)

				for topic := range subscriber.subscriptions {
					subscribers := broker.topics[topic]
					delete(subscribers, subscriber)
				}
			}()

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	broker.accepter = accepter
	return broker, nil
}
