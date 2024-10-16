package Systemge

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Metrics"
)

type SystemgeListener[B any] interface {
	Start() error
	Stop() error

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

	Accept(timeoutNs int64) (SystemgeConnection[B], error)

	GetDefaultCommands() Commands.Handlers

	CheckMetrics() Metrics.MetricsTypes
	GetMetrics() Metrics.MetricsTypes
}

type SystemgeConnection[B any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	/*
		StartReadRoutine(uint32, int64, int64, Tools.ReadHandler[O, C]) error
		StopReadRoutine() error
		IsReadRoutineRunning() bool
	*/

	// SetReadLimit(int64)

	Read(int64) (B, error)
	SetReadDeadline(int64)

	Write(B, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
