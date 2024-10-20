package serviceBroker

import (
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type BrokerSync[D any] struct {
	mutex       sync.RWMutex
	topics      map[string]map[*subscriber[D]]struct{} // topic -> connection -> struct{}
	subscribers map[systemge.Connection[D]]*subscriber[D]
	accepter    *serviceAccepter.Accepter[D]
}

const (
	Subscribe = iota
	Unsubscribe
	Request
	Response
)

type HandleMessageSync[D any] func(
	data D,
	connection systemge.Connection[D],
) (
	messageType uint16,
	topic string,
	payload D,
	syncToken string,
	err error,
)

func NewSync[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterServerConfig *configs.AccepterServer,
	accepterRoutineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	handleAcceptsConcurrently bool,

	readerServerAsyncConfig *configs.ReaderServerAsync,
	readerRoutineConfig *configs.Routine,
	handleReadsConcurrently bool,

	requestResponseManager *tools.RequestResponseManager[D],
	propagateTimeoutNs int64,
	handleMessage HandleMessageSync[D],
	topics []string,
) (*BrokerSync[D], error) {

	broker := &BrokerSync[D]{
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
					messageType, topic, payload, syncToken, err := handleMessage(data, connection)
					if err != nil {
						return
					}

					if messageType == Subscribe {
						broker.mutex.Lock()
						defer broker.mutex.Unlock()

						subscriber, ok := broker.subscribers[connection]
						if !ok {
							return
						}

						subscriber.subscriptions[topic] = struct{}{}

					} else if messageType == Unsubscribe {
						broker.mutex.Lock()
						defer broker.mutex.Unlock()

						subscriber, ok := broker.subscribers[connection]
						if !ok {
							return
						}

						delete(subscriber.subscriptions, topic)
					} else if messageType == Response {
						broker.mutex.Lock()
						defer broker.mutex.Unlock()

						err := requestResponseManager.AddResponse(syncToken, payload)
						if err != nil {
							return
						}
					} else if messageType == Request {
						broker.mutex.RLock()
						defer broker.mutex.RUnlock()

						subscribers, ok := broker.topics[topic]
						if !ok {
							return
						}

						request, err := requestResponseManager.NewRequest(syncToken, responseLimit, timeoutNs)
						if err != nil {
							return
						}
						request.GetResponseChannel()
						go func(request *tools.Request[D]) {
							for {
								select {
								case response, ok := <-request.GetResponseChannel():
									if !ok {
										return
									}
									// propagate sync response
								}
							}
						}(request)

						for subscriber := range subscribers {
							if subscriber.connection == connection {
								continue
							}
							go subscriber.connection.Write(payload, propagateTimeoutNs)
						}
					} else {
						// unknown message type
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
