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
	connection systemge.Connection[D]
	reader     *serviceReader.ReaderAsync[D]
}

func New[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterServerConfig *configs.AccepterServer,
	accepterRoutineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	handleAcceptsConcurrently bool,

	readerServerAsyncConfig *configs.ReaderServerAsync,
	readerRoutineConfig *configs.Routine,
	readHandler tools.ReadHandlerWithError[D, systemge.Connection[D]],
	handleReadsConcurrently bool,

	topics []string,
) (*Broker[D], error) {

	accepterServer, err := serviceAccepter.NewAccepterServer(
		listener,
		accepterServerConfig,
		accepterRoutineConfig,
		acceptHandler,
		handleAcceptsConcurrently,
	)
}
