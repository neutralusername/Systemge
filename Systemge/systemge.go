package Systemge

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Tools"
)

type Listener[B any, C Connection[B]] interface {
	Start() error
	Stop() error

	GetInstanceId() string
	GetSessionId() string
	GetName() string
	GetStatus() int
	GetStopChannel() <-chan struct{}

	Accept(timeoutNs int64) (C, error)

	GetDefaultCommands() Commands.Handlers

	CheckMetrics() Tools.MetricsTypes
	GetMetrics() Tools.MetricsTypes
}

type Connection[B any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	// SetReadLimit(int64) would be nice but redundant for channel communication. integrate this config somehow into tcp and websocket connections

	Read(int64) (B, error)
	SetReadDeadline(int64)

	Write(B, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Tools.MetricsTypes
	CheckMetrics() Tools.MetricsTypes
}
