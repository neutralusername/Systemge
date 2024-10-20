package serviceBroker

import (
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

// propagation based on topic including subscribe and ubsubscribe to particular topics
// and sync request/responses are now two essentially distinct features in this system

type Broker[D any] struct {
	mutex                  sync.RWMutex
	topics                 map[string]map[*subscriber[D]]struct{} // topic -> connection -> struct{}
	subscribers            map[systemge.Connection[D]]*subscriber[D]
	accepter               *serviceAccepter.Accepter[D]
	requestResponseManager *tools.RequestResponseManager[D]
	handleMessage          HandleMessage[D]
	propagateTimeoutNs     int64
}

type subscriber[D any] struct {
	connection    systemge.Connection[D]
	readerAsync   *serviceReader.ReaderAsync[D]
	subscriptions map[string]struct{}
}

const (
	Subscribe = iota
	Unsubscribe
	Response
	Request
	Propagate
)

type HandleMessage[D any] func(
	data D,
	connection systemge.Connection[D],
) (
	messageType uint16,
	topic string,
	payload D,
	syncToken string,
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

	requestResponseManager *tools.RequestResponseManager[D],
	propagateTimeoutNs int64,
	handleMessage HandleMessage[D],
	topics []string,
) (*Broker[D], error) {

	broker := &Broker[D]{
		topics:                 make(map[string]map[*subscriber[D]]struct{}),
		subscribers:            make(map[systemge.Connection[D]]*subscriber[D]),
		requestResponseManager: requestResponseManager,
		handleMessage:          handleMessage,
		propagateTimeoutNs:     propagateTimeoutNs,
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
				broker.readHandler,
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
					connection.Close()
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

func (broker *Broker[D]) readHandler(
	data D,
	connection systemge.Connection[D],
) {
	messageType, topic, payload, syncToken, err := broker.handleMessage(data, connection)
	if err != nil {
		return
	}

	switch messageType {
	case Subscribe:
		broker.mutex.Lock()
		defer broker.mutex.Unlock()

		subscriber, ok := broker.subscribers[connection]
		if !ok {
			return
		}
		subscriber.subscriptions[topic] = struct{}{}

	case Unsubscribe:
		broker.mutex.Lock()
		defer broker.mutex.Unlock()

		subscriber, ok := broker.subscribers[connection]
		if !ok {
			return
		}
		delete(subscriber.subscriptions, topic)

	case Propagate:
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
			go subscriber.connection.Write(payload, broker.propagateTimeoutNs)
		}

	case Request:
		broker.mutex.Lock()
		defer broker.mutex.Unlock()

		_, err := broker.requestResponseManager.NewRequest(
			syncToken,
			responseLimit,
			timeoutNs,
			func(request *tools.Request[D], response D) {
				go connection.Write(response, broker.propagateTimeoutNs)
			},
		)
		if err != nil {
			return
		}

	case Response:
		broker.mutex.Lock()
		defer broker.mutex.Unlock()

		if err := broker.requestResponseManager.AddResponse(syncToken, data); err != nil { // currently has the side effect, that responses to requests may have any topic. might as well be a feature
			return
		}

	default:
		// unknown message type
	}
}
