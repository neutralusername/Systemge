package servicePublishSubscribe

import (
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

// propagation based on topic including subscribe and ubsubscribe to particular topics
// and sync request/responses
// are now two essentially distinct features in this system

type PublishSubscribeServer[D any] struct {
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
) (*PublishSubscribeServer[D], error) {

	publishSubscribeServer := &PublishSubscribeServer[D]{
		topics:                 make(map[string]map[*subscriber[D]]struct{}),
		subscribers:            make(map[systemge.Connection[D]]*subscriber[D]),
		requestResponseManager: requestResponseManager,
		handleMessage:          handleMessage,
		propagateTimeoutNs:     propagateTimeoutNs,
	}
	for _, topic := range topics {
		publishSubscribeServer.topics[topic] = make(map[*subscriber[D]]struct{})
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
				publishSubscribeServer.readHandler,
			)
			if err != nil {
				return err
			}

			subscriber := &subscriber[D]{
				connection:    connection,
				readerAsync:   readerRoutine,
				subscriptions: make(map[string]struct{}),
			}

			publishSubscribeServer.mutex.Lock()
			defer publishSubscribeServer.mutex.Unlock()

			publishSubscribeServer.subscribers[connection] = subscriber

			go func() {
				select {
				case <-connection.GetCloseChannel():
				case <-publishSubscribeServer.accepter.GetRoutine().GetStopChannel():
					connection.Close()
				}

				publishSubscribeServer.mutex.Lock()
				defer publishSubscribeServer.mutex.Unlock()

				subscriber := publishSubscribeServer.subscribers[connection]
				delete(publishSubscribeServer.subscribers, connection)

				for topic := range subscriber.subscriptions {
					subscribers := publishSubscribeServer.topics[topic]
					delete(subscribers, subscriber)
				}
			}()

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	publishSubscribeServer.accepter = accepter
	return publishSubscribeServer, nil
}

func (publishSubscribeServer *PublishSubscribeServer[D]) readHandler(
	data D,
	connection systemge.Connection[D],
) {
	messageType, topic, payload, syncToken, err := publishSubscribeServer.handleMessage(data, connection)
	if err != nil {
		return
	}

	switch messageType {
	case Subscribe:
		publishSubscribeServer.mutex.Lock()
		defer publishSubscribeServer.mutex.Unlock()

		subscriber, ok := publishSubscribeServer.subscribers[connection]
		if !ok {
			return
		}
		subscriber.subscriptions[topic] = struct{}{}

	case Unsubscribe:
		publishSubscribeServer.mutex.Lock()
		defer publishSubscribeServer.mutex.Unlock()

		subscriber, ok := publishSubscribeServer.subscribers[connection]
		if !ok {
			return
		}
		delete(subscriber.subscriptions, topic)

	case Propagate:
		publishSubscribeServer.mutex.RLock()
		defer publishSubscribeServer.mutex.RUnlock()

		subscribers, ok := publishSubscribeServer.topics[topic]
		if !ok {
			return
		}

		for subscriber := range subscribers {
			if subscriber.connection == connection {
				continue
			}
			go subscriber.connection.Write(payload, publishSubscribeServer.propagateTimeoutNs)
		}

	case Request:
		publishSubscribeServer.mutex.Lock()
		defer publishSubscribeServer.mutex.Unlock()

		_, err := publishSubscribeServer.requestResponseManager.NewRequest(
			syncToken,
			responseLimit,
			timeoutNs,
			func(request *tools.Request[D], response D) {
				go connection.Write(response, publishSubscribeServer.propagateTimeoutNs)
			},
		)
		if err != nil {
			return
		}

	case Response:
		publishSubscribeServer.mutex.Lock()
		defer publishSubscribeServer.mutex.Unlock()

		if err := publishSubscribeServer.requestResponseManager.AddResponse(syncToken, data); err != nil { // currently has the side effect, that responses to requests may have any topic. might as well be a feature
			return
		}

	default:
		// unknown message type
	}
}
