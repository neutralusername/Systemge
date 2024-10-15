package SystemgeListener

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type SystemgeListener[B any] interface {
	GetInstanceId() string
	GetSessionId() string
	GetName() string
	GetStatus() int
	GetStopChannel() <-chan struct{}

	/*
	   func (listener *WebsocketListener) IsAcceptRoutineRunning() bool
	   func (listener *WebsocketListener) StartAcceptRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, acceptHandler Tools.AcceptHandler[*ConnectionWebsocket.WebsocketConnection]) error
	   func (listener *WebsocketListener) StopAcceptRoutine(abortOngoingCalls bool) error
	*/

	Accept(timeoutNs int64) (SystemgeConnection.SystemgeConnection[B], error)

	Start() error
	Stop() error

	GetDefaultCommands() Commands.Handlers

	CheckMetrics() Metrics.MetricsTypes
	GetMetrics() Metrics.MetricsTypes
}
