package serviceBroker

import (
	"sync"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/serviceAccepter"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type BrokerServer[D any] struct {
	mutex         sync.RWMutex
	topics        map[string]map[*subscriber[D]]struct{} // topic -> connection -> struct{}
	subscriptions map[*subscriber[D]]map[string]struct{} // connection -> topic -> struct{}
}

type subscriber[D any] struct {
	connection  systemge.Connection[D]
	readRoutine *tools.Routine
}

func NewBrokerServer[D any](
	listener systemge.Listener[D, systemge.Connection[D]],
	accepterServerConfig *configs.AccepterServer,
	accepterRoutineConfig *configs.Routine,
	acceptHandler tools.AcceptHandlerWithError[systemge.Connection[D]],
	handleAcceptsConcurrently bool,

	config *configs.ReaderServerAsync,
	readerRoutineConfig *configs.Routine,
	readHandler tools.ReadHandler[D, systemge.Connection[D]],
	handleReadsConcurrently bool,
) (*BrokerServer[D], error) {

	accepterServer, err := serviceAccepter.NewAccepterServer(
		listener,
		accepterServerConfig,
		accepterRoutineConfig,
		acceptHandler,
		handleAcceptsConcurrently,
	)
}
