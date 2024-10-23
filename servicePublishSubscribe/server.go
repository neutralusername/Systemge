package servicePublishSubscribe

import (
	"errors"
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/serviceReader"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type PublishSubscribeServer[T any] struct {
	config   *configs.PublishSubscribeServer
	listener systemge.Listener[T]

	mutex                  sync.RWMutex
	topics                 map[string]map[*subscriber[T]]struct{} // topic -> connection -> struct{}
	subscribers            map[systemge.Connection[T]]*subscriber[T]
	accepter               *serviceAccepter.Accepter[T]
	requestResponseManager *tools.RequestResponseManager[T]
	handleMessage          HandleMessage[T]
}

type subscriber[T any] struct {
	connection    systemge.Connection[T]
	readerAsync   *serviceReader.ReaderAsync[T]
	subscriptions map[string]struct{}
}

const (
	Subscribe = iota
	Unsubscribe
	Respond
	Request
	Propagate
	RequestAndPropagate
	RespondAndPropagate
)

type HandleMessage[T any] func(
	data T,
	connection systemge.Connection[T],
) (
	messageType uint16,
	topic string,
	payload T,
	syncToken string,
	err error,
)

func New[T any](
	listener systemge.Listener[T],
	requestResponseManager *tools.RequestResponseManager[T],
	publishSubscribeServerConfig *configs.PublishSubscribeServer,
	readerAsyncConfig *configs.ReaderAsync,
	readerRoutineConfig *configs.Routine,
	accepterServerConfig *configs.Accepter,
	accepterRoutineConfig *configs.Routine,
	acceptHandler systemge.AcceptHandlerWithError[T],
	handleMessage HandleMessage[T],
) (*PublishSubscribeServer[T], error) {

	if requestResponseManager == nil { // change this so it may optionally be nil
		return nil, errors.New("requestResponseManager is nil")
	}
	if acceptHandler == nil {
		return nil, errors.New("acceptHandler is nil")
	}
	if readerAsyncConfig == nil {
		return nil, errors.New("readerAsyncConfig is nil")
	}
	if readerRoutineConfig == nil {
		return nil, errors.New("readerRoutineConfig is nil")
	}
	if publishSubscribeServerConfig == nil {
		return nil, errors.New("publishSubscribeServerConfig is nil")
	}
	if handleMessage == nil {
		return nil, errors.New("handleMessage is nil")
	}

	publishSubscribeServer := &PublishSubscribeServer[T]{
		config:                 publishSubscribeServerConfig,
		listener:               listener,
		topics:                 make(map[string]map[*subscriber[T]]struct{}),
		subscribers:            make(map[systemge.Connection[T]]*subscriber[T]),
		requestResponseManager: requestResponseManager,
		handleMessage:          handleMessage,
	}
	for _, topic := range publishSubscribeServerConfig.Topics {
		publishSubscribeServer.topics[topic] = make(map[*subscriber[T]]struct{})
	}

	accepter, err := serviceAccepter.New(
		listener,
		accepterServerConfig,
		accepterRoutineConfig,
		func(connection systemge.Connection[T]) error {
			if err := acceptHandler(connection); err != nil {
				// do smthg with the error
				return nil
			}

			reader, err := serviceReader.NewAsync(
				connection,
				readerAsyncConfig,
				readerRoutineConfig,
				publishSubscribeServer.readHandler,
			)
			if err != nil {
				// do smthg with the error
				return err
			}

			subscriber := &subscriber[T]{
				connection:    connection,
				readerAsync:   reader,
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

				// reader listens on connections' close channel

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
		// do smthg with the error (? accepter failed to start)
		return nil, err
	}

	publishSubscribeServer.accepter = accepter
	return publishSubscribeServer, nil
}

func (publishSubscribeServer *PublishSubscribeServer[T]) readHandler(
	data T,
	connection systemge.Connection[T],
) {
	messageType, topic, payload, syncToken, err := publishSubscribeServer.handleMessage(data, connection)
	if err != nil {
		// do smthg with the error
		return
	}

	switch messageType {
	case Subscribe:
		publishSubscribeServer.mutex.Lock()
		defer publishSubscribeServer.mutex.Unlock()

		subscriber, ok := publishSubscribeServer.subscribers[connection]
		if !ok {
			// do smthg with the error
			return
		}
		subscriber.subscriptions[topic] = struct{}{}

	case Unsubscribe:
		publishSubscribeServer.mutex.Lock()
		defer publishSubscribeServer.mutex.Unlock()

		subscriber, ok := publishSubscribeServer.subscribers[connection]
		if !ok {
			// do smthg with the error
			return
		}
		delete(subscriber.subscriptions, topic)

	case Propagate:
		publishSubscribeServer.Propagate(connection, topic, payload)

	case Request:
		publishSubscribeServer.Request(connection, syncToken)

	case Respond:
		publishSubscribeServer.Respond(syncToken, payload)

	case RequestAndPropagate:
		publishSubscribeServer.Request(connection, syncToken)
		publishSubscribeServer.Propagate(connection, topic, payload)

	case RespondAndPropagate:
		publishSubscribeServer.Respond(syncToken, payload)
		publishSubscribeServer.Propagate(connection, topic, payload)

	default:
		// unknown message type
	}
}

func (publishSubscribeServer *PublishSubscribeServer[T]) Propagate(
	publisher systemge.Connection[T],
	topic string,
	payload T,
) {
	publishSubscribeServer.mutex.RLock()
	defer publishSubscribeServer.mutex.RUnlock()

	subscribers, ok := publishSubscribeServer.topics[topic]
	if !ok {
		// do smthg with the error
		return
	}

	for subscriber := range subscribers {
		if subscriber.connection == publisher {
			continue
		}
		go subscriber.connection.Write(payload, publishSubscribeServer.config.PropagateTimeoutNs)
	}
}

func (publishSubscribeServer *PublishSubscribeServer[T]) Request(
	requester systemge.Connection[T],
	syncToken string,
) {
	if _, err := publishSubscribeServer.requestResponseManager.NewRequest(
		syncToken,
		publishSubscribeServer.config.ResponseLimit,
		publishSubscribeServer.config.RequestTimeoutNs,
		func(request *tools.Request[T], response T) {
			go requester.Write(response, publishSubscribeServer.config.PropagateTimeoutNs)
		},
	); err != nil {
		// do smthg with the error
		return
	}
}

func (publishSubscribeServer *PublishSubscribeServer[T]) Respond(
	syncToken string,
	payload T,
) {
	if err := publishSubscribeServer.requestResponseManager.AddResponse(syncToken, payload); err != nil { // currently has the side effect, that responses to requests may have any topic. might as well be a feature
		// do smthg with the error
		return
	}
}
